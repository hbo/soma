package main

import (
	"crypto/tls"
	"fmt"
	"os"
	"strings"

	"github.com/mjolnir42/soma/internal/adm"
	"github.com/mjolnir42/soma/lib/auth"

	"gopkg.in/resty.v1"

	"github.com/boltdb/bolt"
	"github.com/codegangsta/cli"
	"github.com/peterh/liner"
)

// Client is the resty client that is exported for unknown reasons XXX
var Client *resty.Client

// initCommon provides common startup initialization
func initCommon(c *cli.Context) {
	var (
		err error
		//resp    *resty.Response
		session tls.ClientSessionCache
	)
	if err = configSetup(c); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read the configuration: "+
			"%s\n", err.Error())
		os.Exit(1)
	}

	// setup our REST client
	Client = resty.New().SetRESTMode().
		//SetTimeout(Cfg.Run.TimeoutResty). XXX Bad client setting?
		SetDisableWarn(true).
		SetHeader(`User-Agent`, fmt.Sprintf("%s %s", c.App.Name, c.App.Version)).
		SetHostURL(Cfg.Run.SomaAPI.String())

	if Cfg.Run.SomaAPI.Scheme == `https` {
		session = tls.NewLRUClientSessionCache(64)

		// SetTLSClientConfig replaces, SetRootCertificate updates the
		// tls configuration - option ordering is important
		Client = Client.SetTLSClientConfig(&tls.Config{
			ServerName:         strings.SplitN(Cfg.Run.SomaAPI.Host, `:`, 2)[0],
			ClientSessionCache: session,
			MinVersion:         tls.VersionTLS12,
			MaxVersion:         tls.VersionTLS12,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			},
		}).SetRootCertificate(Cfg.Run.CertPath)
	}

	/*
		// check configured API
		if resp, err = Client.R().Head(`/`); err != nil {
			fmt.Fprintf(os.Stderr, "Error tasting the API endpoint: %s\n",
				err.Error())
		} else if resp.StatusCode() != 204 {
			fmt.Fprintf(os.Stderr, "Error, API Url returned %d instead of 204."+
				" Sure this is SOMA?\n", resp.StatusCode())
			os.Exit(1)
		}

		// check who we talked to
		if resp.Header().Get(`X-Powered-By`) != `SOMA Configuration System` {
			fmt.Fprintf(os.Stderr, `Just FYI, at the end of that API URL`+
				` is not SOMA`)
			os.Exit(1)
		}
	*/

	// embed configuration in boltdb wrapper
	store.Configure(
		Cfg.Run.PathBoltDB,
		os.FileMode(uint32(Cfg.Run.ModeBoltDB)),
		&bolt.Options{Timeout: Cfg.Run.TimeoutBoltDB},
	)

	// configure adm client library
	adm.ConfigureClient(Client)
	adm.ActivateAsyncWait(Cfg.AsyncWait)
	adm.AutomaticJobSave(Cfg.JobSave)
	adm.ConfigureCache(&store)
	adm.ConfigureJSONPostProcessor(Cfg.ProcJSON)
}

// boottime is the pre-run target for bootstrapping SOMA or user
// accounts
func boottime(action cli.ActionFunc) cli.ActionFunc {
	return func(c *cli.Context) error {
		initCommon(c)

		// ensure database content structure is in place
		if err := store.EnsureBuckets(); err != nil {
			fmt.Fprintf(os.Stderr, "Database bucket error: %s\n", err)
			return err
		}

		return action(c)
	}
}

// runtime is the regular pre-run target
func runtime(action cli.ActionFunc) cli.ActionFunc {
	return func(c *cli.Context) error {
		var err error
		var token string
		var cred *auth.Token

		// common initialization
		initCommon(c)

		if !c.GlobalBool(`o`) {
			// ensure database content structure is in place
			if err = store.EnsureBuckets(); err != nil {
				fmt.Fprintf(os.Stderr, "Database bucket error: %s\n", err)
				return err
			}
		}

		if c.GlobalBool(`admin`) {
			Cfg.Auth.User = Cfg.AdminAuth.User
			Cfg.Auth.Pass = Cfg.AdminAuth.Pass
			Cfg.Auth.Token = Cfg.AdminAuth.Token
		}

		// prompt for user
		for Cfg.Auth.User == `` {
			if Cfg.Auth.User, err = adm.Read(`user`); err == liner.ErrPromptAborted {
				os.Exit(0)
			} else if err != nil {
				return err
			}
		}

		// no staticly configured token
		if Cfg.Auth.Token == `` {
		reload:
			// load token from BoltDB
			token, err = store.GetActiveToken(Cfg.Auth.User)
			if err == bolt.ErrBucketNotFound {
				// re-login for logout only if logout --all is called,
				// otherwise set hidden doublelogout flag and skip
				if c.Parent().Args().First() == `logout` && !c.Bool(`all`) {
					c.GlobalSet(`doublelogout`, `true`)
					goto skipForLogout
				}
				// no token in cache
				for Cfg.Auth.Pass == `` {
					if Cfg.Auth.Pass, err = adm.Read(`password`); err == liner.ErrPromptAborted {
						os.Exit(0)
					} else if err != nil {
						return err
					}
				}
				// request new token (validated)
				if cred, err = adm.RequestToken(Client, &auth.Token{
					UserName: Cfg.Auth.User,
					Password: Cfg.Auth.Pass,
				}); err != nil {
					return err
				}
				// save token
				if err = store.SaveToken(
					cred.UserName,
					cred.ValidFrom,
					cred.ExpiresAt,
					cred.Token,
				); err != nil {
					return err
				}
				token = cred.Token
			} else if err != nil {
				return err
			} else {
				// no error, token loaded from cache. Check it has not been
				// invalidated; delete it if it has
				if err := adm.ValidateToken(Client, Cfg.Auth.User, token); err != nil {
					// delete token
					if err = store.DeleteToken(
						Cfg.Auth.User,
						token,
					); err != nil {
						return err
					}
					goto reload
				}
			}
		} else {
			token = Cfg.Auth.Token
		}

	skipForLogout:
		// set token for basic auth
		Client = Client.SetBasicAuth(Cfg.Auth.User, token)

		// run action
		return action(c)
	}
}

// comptime is runtime as a different type
func comptime(completion cli.BashCompleteFunc) cli.BashCompleteFunc {
	return func(c *cli.Context) {
		var err error
		var token string
		var cred *auth.Token

		// common initialization
		initCommon(c)

		if !c.GlobalBool(`o`) {
			// ensure database content structure is in place
			if _ = store.EnsureBuckets(); err != nil {
				fmt.Fprintf(os.Stderr, "Database bucket error: %s\n", err)
				return
			}
		}

		if c.GlobalBool(`admin`) {
			Cfg.Auth.User = Cfg.AdminAuth.User
			Cfg.Auth.Pass = Cfg.AdminAuth.Pass
			Cfg.Auth.Token = Cfg.AdminAuth.Token
		}

		// prompt for user
		for Cfg.Auth.User == `` {
			if Cfg.Auth.User, err = adm.Read(`user`); err == liner.ErrPromptAborted {
				os.Exit(0)
			} else if err != nil {
				return
			}
		}

		// no staticly configured token
		if Cfg.Auth.Token == `` {
		reload:
			// load token from BoltDB
			token, err = store.GetActiveToken(Cfg.Auth.User)
			if err == bolt.ErrBucketNotFound {
				// no token in cache
				for Cfg.Auth.Pass == `` {
					if Cfg.Auth.Pass, err = adm.Read(`password`); err == liner.ErrPromptAborted {
						os.Exit(0)
					} else if err != nil {
						return
					}
				}
				// request new token (validated)
				if cred, err = adm.RequestToken(Client, &auth.Token{
					UserName: Cfg.Auth.User,
					Password: Cfg.Auth.Pass,
				}); err != nil {
					return
				}
				// save token
				if err = store.SaveToken(
					cred.UserName,
					cred.ValidFrom,
					cred.ExpiresAt,
					cred.Token,
				); err != nil {
					return
				}
				token = cred.Token
			} else if err != nil {
				return
			} else {
				// no error, token loaded from cache. Check it has not been
				// invalidated; delete it if it has
				if err := adm.ValidateToken(Client, Cfg.Auth.User, token); err != nil {
					// delete token
					if err = store.DeleteToken(
						Cfg.Auth.User,
						token,
					); err != nil {
						return
					}
					goto reload
				}
			}
		} else {
			token = Cfg.Auth.Token
		}

		// set token for basic auth
		Client = Client.SetBasicAuth(Cfg.Auth.User, token)

		// run action
		completion(c)
	}
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
