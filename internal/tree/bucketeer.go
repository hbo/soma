/*-
 * Copyright (c) 2016, 1&1 Internet SE
 * Copyright (c) 2016, Jörg Pernfuß
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package tree

type Bucketeer interface {
	GetBucket() Receiver
	GetEnvironment() string
	GetRepository() string
	GetRepositoryName() string
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
