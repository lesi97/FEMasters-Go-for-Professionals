package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

type WorkoutStore interface {
	CreateWorkout(*Workout) (*Workout, error)
	GetWorkoutById(int64) (*Workout, error)
	UpdateWorkout(workout *Workout, id int64) error
	DeleteWorkout(int64) error
	GetWorkoutOwner(id int64) (int, error)
}

type PostgresWorkoutStore struct {
	db *sql.DB
}

type WorkoutEntry struct {
	ID              int      `json:"id"`
	ExerciseName    string   `json:"exercise_name"`
	Sets            int      `json:"sets"`
	Reps            *int     `json:"reps"` // Pointer because we want to check if nil as this field is optional
	DurationSeconds *int     `json:"duration_seconds"`
	Weight          *float64 `json:"weight"`
	Notes           string   `json:"notes"`
	OrderIndex      int      `json:"order_index"`
}

type Workout struct {
	ID              int            `json:"id"`
	UserID 			int 			`json:"user_id"`
	Title           string         `json:"title"`
	Description     string         `json:"description"`
	DurationMinutes int            `json:"duration_minutes"`
	CaloriesBurned  int            `json:"calories_burned"`
	Entries         []WorkoutEntry `json:"entries"`
}

type UpdateWorkout struct {
	ID              *int            `json:"id"`
	Title           *string         `json:"title"`
	Description     *string         `json:"description"`
	DurationMinutes *int            `json:"duration_minutes"`
	CaloriesBurned  *int            `json:"calories_burned"`
	Entries         []WorkoutEntry  `json:"entries"`
}

func NewPostgresWorkoutStore(db *sql.DB) *PostgresWorkoutStore {
	return &PostgresWorkoutStore{db: db}
}


func (pg *PostgresWorkoutStore) GetWorkoutById(id int64) (*Workout, error) {
	workout := &Workout{}
	var entriesRaw []byte

	query := `
		SELECT 
			w.id,
			w.title,
			w.description,
			w.duration_minutes,
			w.calories_burned, 
			json_agg(
				json_build_object(
				'id', e.id,
				'exercise_name', e.exercise_name,
				'sets', e.sets,
				'reps', e.reps,
				'duration_seconds', e.duration_seconds,
				'weight', e.weight,
				'notes', e.notes,
				'order_index', e.order_index
				) order by e.order_index
			) as entries 
		FROM workouts w
		JOIN workout_entries e on e.workout_id = w.id
		WHERE w.id = $1
		GROUP BY w.id;
	`

	err := pg.db.QueryRow(query, id).Scan( // pg.db.QueryRow expects at least 1 row returned
		&workout.ID,
		&workout.Title,
		&workout.Description,
		&workout.DurationMinutes,
		&workout.CaloriesBurned,
		&entriesRaw,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no data for workout")
	}

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(entriesRaw, &workout.Entries)
	if err != nil {
		return nil, err
	}

	return workout, nil
}

func (pg *PostgresWorkoutStore) CreateWorkout(workout *Workout) (*Workout, error) {

	tx, err := pg.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO workouts 
			(
			user_id,
			title,
			description, 
			duration_minutes, 
			calories_burned
			)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id;
	`

	err = tx.QueryRow(
		query, 
		workout.UserID,
		workout.Title, 
		workout.Description, 
		workout.DurationMinutes, 
		workout.CaloriesBurned,
	).Scan(&workout.ID)
	if err != nil {
		return nil, err
	}

	for _, entry := range workout.Entries {
		query := `
			INSERT INTO workout_entries 
				(
				workout_id, 
				exercise_name, 
				sets, 
				reps, 
				duration_seconds, 
				weight, 
				notes, 
				order_index
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			RETURNING id;
		`
		err = tx.QueryRow(
			query, 
			workout.ID, 
			entry.ExerciseName, 
			entry.Sets, 
			entry.Reps, 
			entry.DurationSeconds, 
			entry.Weight, 
			entry.Notes, 
			entry.OrderIndex,
		).Scan(&entry.ID)
		if err != nil {
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return workout, nil
}

func (pg *PostgresWorkoutStore) UpdateWorkout(workout *Workout, id int64) error {
	tx, err := pg.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		UPDATE workouts
		SET 
			title = $1,
			description = $2,
			duration_minutes = $3,
			calories_burned = $4
		WHERE id = $5;
	`

	_, err = tx.Exec(query, workout.Title, workout.Description, workout.DurationMinutes, workout.CaloriesBurned, id)
	if err != nil {
		return err
	}

	entriesSql := `
		INSERT INTO workout_entries (
			exercise_name,
			sets,
			reps,
			duration_seconds,
			weight,
			notes,
			order_index,
			id,
			workout_id
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)
		ON CONFLICT (id) DO UPDATE SET
			exercise_name = excluded.exercise_name,
			sets = excluded.sets,
			reps = excluded.reps,
			duration_seconds = excluded.duration_seconds,
			weight = excluded.weight,
			notes = excluded.notes,
			order_index = excluded.order_index
	`

	for _, entries := range workout.Entries {
		_, err := tx.Exec(
			entriesSql,
			entries.ExerciseName,
			entries.Sets,
			entries.Reps,
			entries.DurationSeconds,
			entries.Weight,
			entries.Notes,
			entries.OrderIndex,
			entries.ID,
			id,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (pg *PostgresWorkoutStore) DeleteWorkout(id int64) error {
	query := `DELETE FROM workouts WHERE id = $1;`

	result, err := pg.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (pg *PostgresWorkoutStore) GetWorkoutOwner(workoutID int64) (int, error) {
	var userID int

	query := `SELECT user_id FROM workouts WHERE id = $1;`

	err := pg.db.QueryRow(query, workoutID).Scan(&userID)
	if err != nil {
		return 0, err
	}

	return userID, nil
}
