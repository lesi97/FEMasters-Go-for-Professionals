package store_test

import (
	"database/sql"
	"testing"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/lesi97/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("pgx", "host=localhost user=postgres password=postgres dbname=postgres port=5433 sslmode=disable")
	if err != nil {
		t.Fatalf("opening test db: %v", err) // t.Fatalf has it's own return so not required to add return below
	}

	err = store.Migrate(db, "../../migrations/")
	if err != nil {
		t.Fatalf("migrating test db error: %v", err)
	}

	_, err = db.Exec(`TRUNCATE workouts, workout_entries CASCADE`) // wipe db when running the test so it's always blank
	if err != nil {
		t.Fatalf("truncating tables error: %v", err)
	}

	return db
}

func TestCreateWorkout(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	testStore := store.NewPostgresWorkoutStore(db)

	tests := []struct {
		name 	string
		workout *store.Workout
		wantErr bool
	}{
		{
			name: "valid workout",
			workout: &store.Workout{
				Title: "push day",
				Description: "upper body day",
				DurationMinutes: 60,
				CaloriesBurned: 200,
				Entries: []store.WorkoutEntry{
					{
						ExerciseName: "Bench Press",
						Sets: 3,
						Reps: intPtr(10),
						Weight: floatPtr(135.5),
						Notes: "Warm up properly",
						OrderIndex: 1,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "workout with invalid entries",
			workout: &store.Workout{
				Title: "full body",
				Description: "complete workout",
				DurationMinutes: 90,
				CaloriesBurned: 500,
				Entries: []store.WorkoutEntry{
					{
						ExerciseName: "Plank",
						Sets: 3,
						Reps: intPtr(60),
						Notes: "keep form",
						OrderIndex: 1,
					},
					{
						ExerciseName: "Squats",
						Sets: 4,
						Reps: intPtr(12),
						DurationSeconds: intPtr(60), // In the db we specified can't have both duration and reps so this should fail
						Weight: floatPtr(185.0),
						Notes: "full depth",
						OrderIndex: 2,
					},
				},
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			createdWorkout, err := testStore.CreateWorkout(test.workout)
			if test.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.workout.Title, createdWorkout.Title)
			assert.Equal(t, test.workout.Description, createdWorkout.Description)
			assert.Equal(t, test.workout.DurationMinutes, createdWorkout.DurationMinutes)
			assert.Equal(t, test.workout.CaloriesBurned, createdWorkout.CaloriesBurned)

			retrieved, err := testStore.GetWorkoutById(int64(createdWorkout.ID))
			require.NoError(t, err)

			assert.Equal(t, createdWorkout.ID, retrieved.ID)
			assert.Equal(t, createdWorkout.Title, retrieved.Title)
			assert.Equal(t, createdWorkout.Description, retrieved.Description)
			assert.Equal(t, createdWorkout.DurationMinutes, retrieved.DurationMinutes)
			assert.Equal(t, createdWorkout.CaloriesBurned, retrieved.CaloriesBurned)
			assert.Equal(t, len(test.workout.Entries), len(retrieved.Entries))

			for i, entry := range retrieved.Entries {
				assert.Equal(t, test.workout.Entries[i].ExerciseName, entry.ExerciseName)
				assert.Equal(t, test.workout.Entries[i].Sets, entry.Sets)
				assert.Equal(t, test.workout.Entries[i].Reps, entry.Reps)
				assert.Equal(t, test.workout.Entries[i].DurationSeconds, entry.DurationSeconds)
				assert.Equal(t, test.workout.Entries[i].Weight, entry.Weight)
				assert.Equal(t, test.workout.Entries[i].OrderIndex, entry.OrderIndex)
			}
		})
	}
}

func intPtr(i int) *int {
	return &i
}

func floatPtr(i float64) *float64 {
	return &i
}