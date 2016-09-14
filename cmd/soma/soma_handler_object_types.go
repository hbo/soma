package main

import (
	"database/sql"
	"errors"
	"log"
)

// Message structs
type somaObjectTypeRequest struct {
	action     string
	objectType string
	rename     string
	reply      chan []somaObjectTypeResult
}

type somaObjectTypeResult struct {
	err        error
	objectType string
}

/*  Read Access
 *
 */
type somaObjectTypeReadHandler struct {
	input     chan somaObjectTypeRequest
	shutdown  chan bool
	conn      *sql.DB
	list_stmt *sql.Stmt
	show_stmt *sql.Stmt
}

func (r *somaObjectTypeReadHandler) run() {
	var err error

	r.list_stmt, err = r.conn.Prepare(`
	SELECT object_type
	FROM soma.object_types;
	`)
	if err != nil {
		log.Fatal(err)
	}
	r.show_stmt, err = r.conn.Prepare(`
	SELECT object_type
	FROM soma.object_types
	WHERE object_type = $1;
	`)
	if err != nil {
		log.Fatal(err)
	}
	for {
		select {
		case <-r.shutdown:
			break
		case req := <-r.input:
			go func() {
				r.process(&req)
			}()
		}
	}
}

func (r *somaObjectTypeReadHandler) process(q *somaObjectTypeRequest) {
	var objectType string
	var rows *sql.Rows
	var err error
	result := make([]somaObjectTypeResult, 0)

	switch q.action {
	case "list":
		rows, err = r.list_stmt.Query()
		defer rows.Close()
		if err != nil {
			result = append(result, somaObjectTypeResult{
				err:        err,
				objectType: q.objectType,
			})
			q.reply <- result
			return
		}

		for rows.Next() {
			err = rows.Scan(&objectType)
			if err != nil {
				result = append(result, somaObjectTypeResult{
					err:        err,
					objectType: q.objectType,
				})
				err = nil
				continue
			}
			result = append(result, somaObjectTypeResult{
				err:        nil,
				objectType: objectType,
			})
		}
	case "show":
		err = r.show_stmt.QueryRow(q.objectType).Scan(&objectType)
		if err != nil {
			result = append(result, somaObjectTypeResult{
				err:        err,
				objectType: q.objectType,
			})
			q.reply <- result
			return
		}

		result = append(result, somaObjectTypeResult{
			err:        nil,
			objectType: objectType,
		})
	default:
		result = append(result, somaObjectTypeResult{
			err:        errors.New("not implemented"),
			objectType: "",
		})
	}
	q.reply <- result
}

/*
 * Write Access
 */

type somaObjectTypeWriteHandler struct {
	input    chan somaObjectTypeRequest
	shutdown chan bool
	conn     *sql.DB
	add_stmt *sql.Stmt
	del_stmt *sql.Stmt
	ren_stmt *sql.Stmt
}

func (w *somaObjectTypeWriteHandler) run() {
	var err error

	w.add_stmt, err = w.conn.Prepare(`
  INSERT INTO soma.object_types (object_type)
  SELECT $1 WHERE NOT EXISTS (
    SELECT object_type
	FROM soma.object_types
	WHERE object_type = $2
  );
  `)
	if err != nil {
		log.Fatal(err)
	}
	defer w.add_stmt.Close()

	w.del_stmt, err = w.conn.Prepare(`
  DELETE FROM soma.object_types
  WHERE object_type = $1;
  `)
	if err != nil {
		log.Fatal(err)
	}
	defer w.del_stmt.Close()

	w.ren_stmt, err = w.conn.Prepare(`
  UPDATE soma.object_types
  SET object_type = $1
  WHERE object_type = $2;
  `)
	if err != nil {
		log.Fatal(err)
	}
	defer w.ren_stmt.Close()

	for {
		select {
		case <-w.shutdown:
			break
		case req := <-w.input:
			w.process(&req)
		}
	}
}

func (w *somaObjectTypeWriteHandler) process(q *somaObjectTypeRequest) {
	var res sql.Result
	var err error

	result := make([]somaObjectTypeResult, 0)
	switch q.action {
	case "add":
		res, err = w.add_stmt.Exec(q.objectType, q.objectType)
	case "delete":
		res, err = w.del_stmt.Exec(q.objectType)
	case "rename":
		res, err = w.ren_stmt.Exec(q.rename, q.objectType)
	default:
		result = append(result, somaObjectTypeResult{
			err:        errors.New("not implemented"),
			objectType: "",
		})
		q.reply <- result
		return
	}
	if err != nil {
		result = append(result, somaObjectTypeResult{
			err:        err,
			objectType: q.objectType,
		})
		q.reply <- result
		return
	}

	rowCnt, _ := res.RowsAffected()
	if rowCnt == 0 {
		result = append(result, somaObjectTypeResult{
			err:        errors.New("No rows affected"),
			objectType: q.objectType,
		})
	} else if rowCnt > 1 {
		result = append(result, somaObjectTypeResult{
			err:        errors.New("Too many rows affected"),
			objectType: q.objectType,
		})
	} else {
		result = append(result, somaObjectTypeResult{
			err:        nil,
			objectType: q.objectType,
		})
	}
	q.reply <- result
}

/* Ops Access
 */
func (r *somaObjectTypeReadHandler) shutdownNow() {
	r.shutdown <- true
}

func (w *somaObjectTypeWriteHandler) shutdownNow() {
	w.shutdown <- true
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix