

# SQL Gen
[![Build Status](https://travis-ci.org/natebrennand/sql_gen.svg)](https://travis-ci.org/natebrennand/sql_gen)


SQL Gen creates basic structs and methods based on a SQL schema passed in to assist development. Currently there is only Postgres support but MySQL is on the roadmap.
Schema's are read through STDIN.

## Generating your schema

#### Postgres

```bash
pg_dump -s yourdatabasename > schema.sql
```

## Results

```bash
pg_dump -s test > schema.sql
./sql_gen < schema.sql

# alternatively

./sql_gen < pg_dump -s test
```

Input schema:
```sql
CREATE TABLE test (
    term character varying(32),
    callnumber integer,
    classnotes character varying(64),
    starttime1 time without time zone,
    description text
);
```

Resulting code:
```go
package main

import (
	"database/sql"
	"fmt"
	"time"
)

type test struct {
	Term        string
	Callnumber  int
	Classnotes  string
	Starttime1  time.Time
	Description string
}

func (t *test) Scan(row *sql.Row) error {
	return row.Scan(
		&t.Term,
		&t.Callnumber,
		&t.Classnotes,
		&t.Starttime1,
		&t.Description,
	)
}

func (t test) Insert(db *sql.DB) error {
	query := "INSERT INTO test (term, callnumber, classnotes, starttime1, description) VALUES ($1, $2, $3, $4, $5)"
	_, err := db.Exec(
		query,
		t.Term,
		t.Callnumber,
		t.Classnotes,
		t.Starttime1.Format("15:04"),
		t.Description,
	)
	if err != nil {
		return fmt.Errorf("Failed to insert Course, %#v, => %s", t, err.Error())
	}
	return nil
}
```
