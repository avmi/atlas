// Copyright 2021-present The Atlas Authors. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

//go:build !ent

package postgres

import (
	"context"
	"fmt"

	"ariga.io/atlas/schemahcl"
	"ariga.io/atlas/sql/internal/specutil"
	"ariga.io/atlas/sql/internal/sqlx"
	"ariga.io/atlas/sql/migrate"
	"ariga.io/atlas/sql/schema"
	"ariga.io/atlas/sql/sqlspec"
)

var (
	specOptions []schemahcl.Option
	specFuncs   = &specutil.SchemaFuncs{
		Table: tableSpec,
		View:  viewSpec,
	}
	scanFuncs = &specutil.ScanFuncs{
		Table: convertTable,
		View:  convertView,
	}
)

func tableAttrsSpec(*schema.Table, *sqlspec.Table) {
	// unimplemented.
}

func convertTableAttrs(*sqlspec.Table, *schema.Table) error {
	return nil // unimplemented.
}

// tableAttrDiff allows extending table attributes diffing with build-specific logic.
func (*diff) tableAttrDiff(_, _ *schema.Table) ([]schema.Change, error) {
	return nil, nil // unimplemented.
}

// addTableAttrs allows extending table attributes creation with build-specific logic.
func (*state) addTableAttrs(_ *schema.AddTable) {
	// unimplemented.
}

// alterTableAttr allows extending table attributes alteration with build-specific logic.
func (s *state) alterTableAttr(*sqlx.Builder, *schema.ModifyAttr) {
	// unimplemented.
}

func realmObjectsSpec(*doc, *schema.Realm) error {
	return nil // unimplemented.
}

func triggersSpec([]*schema.Trigger, *doc) error {
	return nil // unimplemented.
}

func (*inspect) inspectViews(context.Context, *schema.Realm, *schema.InspectOptions) error {
	return nil // unimplemented.
}

func (*inspect) inspectFuncs(context.Context, *schema.Realm, *schema.InspectOptions) error {
	return nil // unimplemented.
}

func (*inspect) inspectTypes(context.Context, *schema.Realm, *schema.InspectOptions) error {
	return nil // unimplemented.
}

func (*inspect) inspectObjects(context.Context, *schema.Realm, *schema.InspectOptions) error {
	return nil // unimplemented.
}

func (*inspect) inspectTriggers(context.Context, *schema.Realm, *schema.InspectOptions) error {
	return nil // unimplemented.
}

func (*inspect) inspectDeps(context.Context, *schema.Realm, *schema.InspectOptions) error {
	return nil // unimplemented.
}

func (*inspect) inspectRealmObjects(context.Context, *schema.Realm, *schema.InspectOptions) error {
	return nil // unimplemented.
}

func (*state) addView(*schema.AddView) error {
	return nil // unimplemented.
}

func (*state) dropView(*schema.DropView) error {
	return nil // unimplemented.
}

func (*state) modifyView(*schema.ModifyView) error {
	return nil // unimplemented.
}

func (*state) renameView(*schema.RenameView) {
	// unimplemented.
}

func (s *state) addFunc(*schema.AddFunc) error {
	return nil // unimplemented.
}

func (s *state) dropFunc(*schema.DropFunc) error {
	return nil // unimplemented.
}

func (s *state) modifyFunc(*schema.ModifyFunc) error {
	return nil // unimplemented.
}

func (s *state) renameFunc(*schema.RenameFunc) error {
	return nil // unimplemented.
}

func (s *state) addProc(*schema.AddProc) error {
	return nil // unimplemented.
}

func (s *state) dropProc(*schema.DropProc) error {
	return nil // unimplemented.
}

func (s *state) modifyProc(*schema.ModifyProc) error {
	return nil // unimplemented.
}

func (s *state) renameProc(*schema.RenameProc) error {
	return nil // unimplemented.
}

func (s *state) addObject(add *schema.AddObject) error {
	switch o := add.O.(type) {
	case *schema.EnumType:
		create, drop := s.createDropEnum(o)
		s.append(&migrate.Change{
			Source:  add,
			Cmd:     create,
			Reverse: drop,
			Comment: fmt.Sprintf("create enum type %q", o.T),
		})
	default:
		// unsupported object type.
	}
	return nil
}

func (s *state) dropObject(drop *schema.DropObject) error {
	switch o := drop.O.(type) {
	case *schema.EnumType:
		create, dropE := s.createDropEnum(o)
		s.append(&migrate.Change{
			Source:  drop,
			Cmd:     dropE,
			Reverse: create,
			Comment: fmt.Sprintf("drop enum type %q", o.T),
		})
	default:
		// unsupported object type.
	}
	return nil
}

func (s *state) modifyObject(modify *schema.ModifyObject) error {
	if _, ok := modify.From.(*schema.EnumType); ok {
		return s.alterEnum(modify)
	}
	return nil // unimplemented.
}

func (*state) addTrigger(*schema.AddTrigger) error {
	return nil // unimplemented.
}

func (*state) dropTrigger(*schema.DropTrigger) error {
	return nil // unimplemented.
}

func (*state) renameTrigger(*schema.RenameTrigger) error {
	return nil // unimplemented.
}

func (*state) modifyTrigger(*schema.ModifyTrigger) error {
	return nil // unimplemented.
}

func (*diff) ViewAttrChanges(_, _ *schema.View) []schema.Change {
	return nil // unimplemented.
}

// RealmObjectDiff returns a changeset for migrating realm (database) objects
// from one state to the other. For example, adding extensions or users.
func (*diff) RealmObjectDiff(_, _ *schema.Realm) ([]schema.Change, error) {
	return nil, nil // unimplemented.
}

// SchemaObjectDiff returns a changeset for migrating schema objects from
// one state to the other.
func (*diff) SchemaObjectDiff(from, to *schema.Schema, _ *schema.DiffOptions) ([]schema.Change, error) {
	var changes []schema.Change
	// Drop or modify enums.
	for _, o1 := range from.Objects {
		e1, ok := o1.(*schema.EnumType)
		if !ok {
			continue // Unsupported object type.
		}
		o2, ok := to.Object(func(o schema.Object) bool {
			e2, ok := o.(*schema.EnumType)
			return ok && e1.T == e2.T
		})
		if !ok {
			changes = append(changes, &schema.DropObject{O: o1})
			continue
		}
		if e2 := o2.(*schema.EnumType); !sqlx.ValuesEqual(e1.Values, e2.Values) {
			changes = append(changes, &schema.ModifyObject{From: e1, To: e2})
		}
	}
	// Add new enums.
	for _, o1 := range to.Objects {
		e1, ok := o1.(*schema.EnumType)
		if !ok {
			continue // Unsupported object type.
		}
		if _, ok := from.Object(func(o schema.Object) bool {
			e2, ok := o.(*schema.EnumType)
			return ok && e1.T == e2.T
		}); !ok {
			changes = append(changes, &schema.AddObject{O: e1})
		}
	}
	return changes, nil
}

func verifyChanges(context.Context, []schema.Change) error {
	return nil // unimplemented.
}

func convertDomains(_ []*sqlspec.Table, domains []*domain, _ *schema.Realm) error {
	if len(domains) > 0 {
		return fmt.Errorf("postgres: domains are not supported by this version. Use: https://atlasgo.io/getting-started")
	}
	return nil
}

func convertAggregate(d *doc, _ *schema.Realm) error {
	if len(d.Aggregates) > 0 {
		return fmt.Errorf("postgres: aggregates are not supported by this version. Use: https://atlasgo.io/getting-started")
	}
	return nil
}

func convertSequences(_ []*sqlspec.Table, seqs []*sqlspec.Sequence, _ *schema.Realm) error {
	if len(seqs) > 0 {
		return fmt.Errorf("postgres: sequences are not supported by this version. Use: https://atlasgo.io/getting-started")
	}
	return nil
}

func convertPolicies(_ []*sqlspec.Table, ps []*policy, _ *schema.Realm) error {
	if len(ps) > 0 {
		return fmt.Errorf("postgres: policies are not supported by this version. Use: https://atlasgo.io/getting-started")
	}
	return nil
}

func convertExtensions(exs []*extension, _ *schema.Realm) error {
	if len(exs) > 0 {
		return fmt.Errorf("postgres: extensions are not supported by this version. Use: https://atlasgo.io/getting-started")
	}
	return nil
}

func convertEventTriggers(evs []*eventTrigger, _ *schema.Realm) error {
	if len(evs) > 0 {
		return fmt.Errorf("postgres: event triggers are not supported by this version. Use: https://atlasgo.io/getting-started")
	}
	return nil
}

func normalizeRealm(*schema.Realm) error {
	return nil
}

func schemasObjectSpec(*doc, ...*schema.Schema) error {
	return nil // unimplemented.
}

// objectSpec converts from a concrete schema objects into specs.
func objectSpec(d *doc, spec *specutil.SchemaSpec, s *schema.Schema) error {
	for _, o := range s.Objects {
		if e, ok := o.(*schema.EnumType); ok {
			d.Enums = append(d.Enums, &enum{
				Name:   e.T,
				Values: e.Values,
				Schema: specutil.SchemaRef(spec.Schema.Name),
			})
		}
	}
	return nil
}

// convertEnums converts possibly referenced column types (like enums) to
// an actual schema.Type and sets it on the correct schema.Column.
func convertTypes(d *doc, r *schema.Realm) error {
	if len(d.Enums) == 0 {
		return nil
	}
	byName := make(map[string]*schema.EnumType)
	for _, e := range d.Enums {
		if byName[e.Name] != nil {
			return fmt.Errorf("duplicate enum %q", e.Name)
		}
		ns, err := specutil.SchemaName(e.Schema)
		if err != nil {
			return fmt.Errorf("extract schema name from enum reference: %w", err)
		}
		es, ok := r.Schema(ns)
		if !ok {
			return fmt.Errorf("schema %q defined on enum %q was not found in realm", ns, e.Name)
		}
		e1 := &schema.EnumType{T: e.Name, Schema: es, Values: e.Values}
		es.AddObjects(e1)
		byName[e.Name] = e1
	}
	for _, t := range d.Tables {
		for _, c := range t.Columns {
			var enum *schema.EnumType
			switch {
			case c.Type.IsRefTo("enum"):
				n, err := enumName(c.Type)
				if err != nil {
					return err
				}
				e, ok := byName[n]
				if !ok {
					return fmt.Errorf("enum %q was not found in realm", n)
				}
				enum = e
			default:
				if n, ok := arrayType(c.Type.T); ok {
					enum = byName[n]
				}
			}
			if enum == nil {
				continue
			}
			schemaT, err := specutil.SchemaName(t.Schema)
			if err != nil {
				return fmt.Errorf("extract schema name from table reference: %w", err)
			}
			ts, ok := r.Schema(schemaT)
			if !ok {
				return fmt.Errorf("schema %q not found in realm for table %q", schemaT, t.Name)
			}
			tt, ok := ts.Table(t.Name)
			if !ok {
				return fmt.Errorf("table %q not found in schema %q", t.Name, ts.Name)
			}
			cc, ok := tt.Column(c.Name)
			if !ok {
				return fmt.Errorf("column %q not found in table %q", c.Name, t.Name)
			}
			switch t := cc.Type.Type.(type) {
			case *ArrayType:
				t.Type = enum
			default:
				cc.Type.Type = enum
			}
		}
	}
	return nil
}

func indexToUnique(*schema.ModifyIndex) (*AddUniqueConstraint, bool) {
	return nil, false // unimplemented.
}

func uniqueConstChanged(_, _ []schema.Attr) bool {
	// Unsupported change in package mode (ariga.io/sql/postgres)
	// to keep BC with old versions.
	return false
}

func excludeConstChanged(_, _ []schema.Attr) bool {
	// Unsupported change in package mode (ariga.io/sql/postgres)
	// to keep BC with old versions.
	return false
}

func convertExclude(schemahcl.Resource, *schema.Table) error {
	return nil // unimplemented.
}

func (*state) sortChanges(changes []schema.Change) []schema.Change {
	return sqlx.SortChanges(changes, nil)
}

func (*state) detachCycles(changes []schema.Change) ([]schema.Change, error) {
	return sqlx.DetachCycles(changes)
}

func excludeSpec(*sqlspec.Table, *sqlspec.Index, *schema.Index, *Constraint) error {
	return nil // unimplemented.
}

const (
	// Query to list tables information.
	// Note, 'attrs' are not supported in this version.
	tablesQuery = `
SELECT
	t3.oid,
	t1.table_schema,
	t1.table_name,
	pg_catalog.obj_description(t3.oid, 'pg_class') AS comment,
	t4.partattrs AS partition_attrs,
	t4.partstrat AS partition_strategy,
	pg_get_expr(t4.partexprs, t4.partrelid) AS partition_exprs,
	'{}' AS attrs
FROM
	INFORMATION_SCHEMA.TABLES AS t1
	JOIN pg_catalog.pg_namespace AS t2 ON t2.nspname = t1.table_schema
	JOIN pg_catalog.pg_class AS t3 ON t3.relnamespace = t2.oid AND t3.relname = t1.table_name
	LEFT JOIN pg_catalog.pg_partitioned_table AS t4 ON t4.partrelid = t3.oid
	LEFT JOIN pg_depend AS t5 ON t5.classid = 'pg_catalog.pg_class'::regclass::oid AND t5.objid = t3.oid AND t5.deptype = 'e'
WHERE
	t1.table_type = 'BASE TABLE'
	AND NOT COALESCE(t3.relispartition, false)
	AND t1.table_schema IN (%s)
	AND t5.objid IS NULL
ORDER BY
	t1.table_schema, t1.table_name
`
	// Query to list tables by their names.
	// Note, 'attrs' are not supported in this version.
	tablesQueryArgs = `
SELECT
	t3.oid,
	t1.table_schema,
	t1.table_name,
	pg_catalog.obj_description(t3.oid, 'pg_class') AS comment,
	t4.partattrs AS partition_attrs,
	t4.partstrat AS partition_strategy,
	pg_get_expr(t4.partexprs, t4.partrelid) AS partition_exprs,
	'{}' AS attrs
FROM
	INFORMATION_SCHEMA.TABLES AS t1
	JOIN pg_catalog.pg_namespace AS t2 ON t2.nspname = t1.table_schema
	JOIN pg_catalog.pg_class AS t3 ON t3.relnamespace = t2.oid AND t3.relname = t1.table_name
	LEFT JOIN pg_catalog.pg_partitioned_table AS t4 ON t4.partrelid = t3.oid
	LEFT JOIN pg_depend AS t5 ON t5.classid = 'pg_catalog.pg_class'::regclass::oid AND t5.objid = t3.oid AND t5.deptype = 'e'
WHERE
	t1.table_type = 'BASE TABLE'
	AND NOT COALESCE(t3.relispartition, false)
	AND t1.table_schema IN (%s)
	AND t1.table_name IN (%s)
	AND t5.objid IS NULL
ORDER BY
	t1.table_schema, t1.table_name
`
)
