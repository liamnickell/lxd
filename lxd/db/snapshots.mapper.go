//go:build linux && cgo && !agent
// +build linux,cgo,!agent

package db

// The code below was generated by lxd-generate - DO NOT EDIT!

import (
	"database/sql"
	"fmt"

	"github.com/lxc/lxd/lxd/db/cluster"
	"github.com/lxc/lxd/lxd/db/query"
	"github.com/lxc/lxd/shared/api"
)

var _ = api.ServerEnvironment{}

var instanceSnapshotObjects = cluster.RegisterStmt(`
SELECT instances_snapshots.id, projects.name AS project, instances.name AS instance, instances_snapshots.name, instances_snapshots.creation_date, instances_snapshots.stateful, coalesce(instances_snapshots.description, ''), instances_snapshots.expiry_date
  FROM instances_snapshots JOIN projects ON instances.project_id = projects.id JOIN instances ON instances_snapshots.instance_id = instances.id
  ORDER BY projects.id, instances.id, instances_snapshots.name
`)

var instanceSnapshotObjectsByProjectAndInstance = cluster.RegisterStmt(`
SELECT instances_snapshots.id, projects.name AS project, instances.name AS instance, instances_snapshots.name, instances_snapshots.creation_date, instances_snapshots.stateful, coalesce(instances_snapshots.description, ''), instances_snapshots.expiry_date
  FROM instances_snapshots JOIN projects ON instances.project_id = projects.id JOIN instances ON instances_snapshots.instance_id = instances.id
  WHERE project = ? AND instance = ? ORDER BY projects.id, instances.id, instances_snapshots.name
`)

var instanceSnapshotObjectsByProjectAndInstanceAndName = cluster.RegisterStmt(`
SELECT instances_snapshots.id, projects.name AS project, instances.name AS instance, instances_snapshots.name, instances_snapshots.creation_date, instances_snapshots.stateful, coalesce(instances_snapshots.description, ''), instances_snapshots.expiry_date
  FROM instances_snapshots JOIN projects ON instances.project_id = projects.id JOIN instances ON instances_snapshots.instance_id = instances.id
  WHERE project = ? AND instance = ? AND instances_snapshots.name = ? ORDER BY projects.id, instances.id, instances_snapshots.name
`)

var instanceSnapshotID = cluster.RegisterStmt(`
SELECT instances_snapshots.id FROM instances_snapshots JOIN projects ON instances.project_id = projects.id JOIN instances ON instances_snapshots.instance_id = instances.id
  WHERE projects.name = ? AND instances.name = ? AND instances_snapshots.name = ?
`)

var instanceSnapshotCreate = cluster.RegisterStmt(`
INSERT INTO instances_snapshots (instance_id, name, creation_date, stateful, description, expiry_date)
  VALUES ((SELECT instances.id FROM instances JOIN projects ON projects.id = instances.project_id WHERE projects.name = ? AND instances.name = ?), ?, ?, ?, ?, ?)
`)

var instanceSnapshotRename = cluster.RegisterStmt(`
UPDATE instances_snapshots SET name = ? WHERE instance_id = (SELECT instances.id FROM instances JOIN projects ON projects.id = instances.project_id WHERE projects.name = ? AND instances.name = ?) AND name = ?
`)

var instanceSnapshotDeleteByProjectAndInstanceAndName = cluster.RegisterStmt(`
DELETE FROM instances_snapshots WHERE instance_id = (SELECT instances.id FROM instances JOIN projects ON projects.id = instances.project_id WHERE projects.name = ? AND instances.name = ?) AND name = ?
`)

// GetInstanceSnapshots returns all available instance_snapshots.
// generator: instance_snapshot GetMany
func (c *ClusterTx) GetInstanceSnapshots(filter InstanceSnapshotFilter) ([]InstanceSnapshot, error) {
	var err error

	// Result slice.
	objects := make([]InstanceSnapshot, 0)

	// Pick the prepared statement and arguments to use based on active criteria.
	var stmt *sql.Stmt
	var args []interface{}

	if filter.Project != nil && filter.Instance != nil && filter.Name != nil {
		stmt = c.stmt(instanceSnapshotObjectsByProjectAndInstanceAndName)
		args = []interface{}{
			filter.Project,
			filter.Instance,
			filter.Name,
		}
	} else if filter.Project != nil && filter.Instance != nil && filter.Name == nil {
		stmt = c.stmt(instanceSnapshotObjectsByProjectAndInstance)
		args = []interface{}{
			filter.Project,
			filter.Instance,
		}
	} else if filter.Project == nil && filter.Instance == nil && filter.Name == nil {
		stmt = c.stmt(instanceSnapshotObjects)
		args = []interface{}{}
	} else {
		return nil, fmt.Errorf("No statement exists for the given Filter")
	}

	// Dest function for scanning a row.
	dest := func(i int) []interface{} {
		objects = append(objects, InstanceSnapshot{})
		return []interface{}{
			&objects[i].ID,
			&objects[i].Project,
			&objects[i].Instance,
			&objects[i].Name,
			&objects[i].CreationDate,
			&objects[i].Stateful,
			&objects[i].Description,
			&objects[i].ExpiryDate,
		}
	}

	// Select.
	err = query.SelectObjects(stmt, dest, args...)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch from \"instances_snapshots\" table: %w", err)
	}

	config, err := c.GetConfig("instance_snapshot")
	if err != nil {
		return nil, err
	}

	for i := range objects {
		if _, ok := config[objects[i].ID]; !ok {
			objects[i].Config = map[string]string{}
		} else {
			objects[i].Config = config[objects[i].ID]
		}
	}

	devices, err := c.GetDevices("instance_snapshot")
	if err != nil {
		return nil, err
	}

	for i := range objects {
		objects[i].Devices = map[string]Device{}
		for _, obj := range devices[objects[i].ID] {
			if _, ok := objects[i].Devices[obj.Name]; !ok {
				objects[i].Devices[obj.Name] = obj
			} else {
				return nil, fmt.Errorf("Found duplicate Device with name %q", obj.Name)
			}
		}
	}

	return objects, nil
}

// GetInstanceSnapshot returns the instance_snapshot with the given key.
// generator: instance_snapshot GetOne
func (c *ClusterTx) GetInstanceSnapshot(project string, instance string, name string) (*InstanceSnapshot, error) {
	filter := InstanceSnapshotFilter{}
	filter.Project = &project
	filter.Instance = &instance
	filter.Name = &name

	objects, err := c.GetInstanceSnapshots(filter)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch from \"instances_snapshots\" table: %w", err)
	}

	switch len(objects) {
	case 0:
		return nil, ErrNoSuchObject
	case 1:
		return &objects[0], nil
	default:
		return nil, fmt.Errorf("More than one \"instances_snapshots\" entry matches")
	}
}

// GetInstanceSnapshotID return the ID of the instance_snapshot with the given key.
// generator: instance_snapshot ID
func (c *ClusterTx) GetInstanceSnapshotID(project string, instance string, name string) (int64, error) {
	stmt := c.stmt(instanceSnapshotID)
	rows, err := stmt.Query(project, instance, name)
	if err != nil {
		return -1, fmt.Errorf("Failed to get \"instances_snapshots\" ID: %w", err)
	}

	defer rows.Close()

	// Ensure we read one and only one row.
	if !rows.Next() {
		return -1, ErrNoSuchObject
	}
	var id int64
	err = rows.Scan(&id)
	if err != nil {
		return -1, fmt.Errorf("Failed to scan ID: %w", err)
	}

	if rows.Next() {
		return -1, fmt.Errorf("More than one row returned")
	}
	err = rows.Err()
	if err != nil {
		return -1, fmt.Errorf("Result set failure: %w", err)
	}

	return id, nil
}

// InstanceSnapshotExists checks if a instance_snapshot with the given key exists.
// generator: instance_snapshot Exists
func (c *ClusterTx) InstanceSnapshotExists(project string, instance string, name string) (bool, error) {
	_, err := c.GetInstanceSnapshotID(project, instance, name)
	if err != nil {
		if err == ErrNoSuchObject {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// CreateInstanceSnapshot adds a new instance_snapshot to the database.
// generator: instance_snapshot Create
func (c *ClusterTx) CreateInstanceSnapshot(object InstanceSnapshot) (int64, error) {
	// Check if a instance_snapshot with the same key exists.
	exists, err := c.InstanceSnapshotExists(object.Project, object.Instance, object.Name)
	if err != nil {
		return -1, fmt.Errorf("Failed to check for duplicates: %w", err)
	}

	if exists {
		return -1, fmt.Errorf("This \"instances_snapshots\" entry already exists")
	}

	args := make([]interface{}, 7)

	// Populate the statement arguments.
	args[0] = object.Project
	args[1] = object.Instance
	args[2] = object.Name
	args[3] = object.CreationDate
	args[4] = object.Stateful
	args[5] = object.Description
	args[6] = object.ExpiryDate

	// Prepared statement to use.
	stmt := c.stmt(instanceSnapshotCreate)

	// Execute the statement.
	result, err := stmt.Exec(args...)
	if err != nil {
		return -1, fmt.Errorf("Failed to create \"instances_snapshots\" entry: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return -1, fmt.Errorf("Failed to fetch \"instances_snapshots\" entry ID: %w", err)
	}

	referenceID := int(id)
	for key, value := range object.Config {
		insert := Config{
			ReferenceID: referenceID,
			Key:         key,
			Value:       value,
		}

		err = c.CreateConfig("instance_snapshot", insert)
		if err != nil {
			return -1, fmt.Errorf("Insert Config failed for InstanceSnapshot: %w", err)
		}

	}
	for _, insert := range object.Devices {
		insert.ReferenceID = int(id)
		err = c.CreateDevice("instance_snapshot", insert)
		if err != nil {
			return -1, fmt.Errorf("Insert Devices failed for InstanceSnapshot: %w", err)
		}

	}
	return id, nil
}

// RenameInstanceSnapshot renames the instance_snapshot matching the given key parameters.
// generator: instance_snapshot Rename
func (c *ClusterTx) RenameInstanceSnapshot(project string, instance string, name string, to string) error {
	stmt := c.stmt(instanceSnapshotRename)
	result, err := stmt.Exec(to, project, instance, name)
	if err != nil {
		return fmt.Errorf("Rename InstanceSnapshot failed: %w", err)
	}

	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("Fetch affected rows failed: %w", err)
	}

	if n != 1 {
		return fmt.Errorf("Query affected %d rows instead of 1", n)
	}
	return nil
}

// DeleteInstanceSnapshot deletes the instance_snapshot matching the given key parameters.
// generator: instance_snapshot DeleteOne-by-Project-and-Instance-and-Name
func (c *ClusterTx) DeleteInstanceSnapshot(project string, instance string, name string) error {
	stmt := c.stmt(instanceSnapshotDeleteByProjectAndInstanceAndName)
	result, err := stmt.Exec(project, instance, name)
	if err != nil {
		return fmt.Errorf("Delete \"instances_snapshots\": %w", err)
	}

	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("Fetch affected rows: %w", err)
	}

	if n != 1 {
		return fmt.Errorf("Query deleted %d rows instead of 1", n)
	}

	return nil
}
