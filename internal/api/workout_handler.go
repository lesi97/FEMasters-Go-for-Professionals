package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/lesi97/internal/store"
)

type WorkoutHandler struct {
	workoutStore store.WorkoutStore
}

func NewWorkoutHandler(workoutStore store.WorkoutStore) *WorkoutHandler {
	return &WorkoutHandler{
		workoutStore: workoutStore,
	}
}

func (wh *WorkoutHandler) HandleGetWorkoutById(w http.ResponseWriter, r *http.Request) {
	paramsWorkoutId := chi.URLParam(r, "id")
	if paramsWorkoutId == "" {
		fmt.Println("No ID provided")
		http.Error(w, "No ID provided", http.StatusInternalServerError)
		return
	}
	
	workoutId, err := strconv.ParseInt(paramsWorkoutId, 10, 64) // Base 10, 64 bit int
	if err != nil {
		fmt.Println(err)
		http.NotFound(w, r)
		return
	}

	workout, err := wh.workoutStore.GetWorkoutById(workoutId)
	if err != nil {
		fmt.Println(err)
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(workout)

}

func (wh *WorkoutHandler) HandleCreateWorkout(w http.ResponseWriter, r *http.Request) {
	var workout store.Workout

	err := json.NewDecoder(r.Body).Decode(&workout) // Use NewDecoder when accepting JSON from HTTP, Unmarshal for internal JSON
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to decode JSON", http.StatusInternalServerError)
		return
	}

	createdWorkout, err := wh.workoutStore.CreateWorkout(&workout)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to create workout", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(createdWorkout)
}

func (wh *WorkoutHandler) HandleUpdateWorkout(w http.ResponseWriter, r *http.Request) {
	paramsWorkoutId := chi.URLParam(r, "id")
	if paramsWorkoutId == "" {
		fmt.Println("No ID provided")
		http.Error(w, "No ID provided", http.StatusInternalServerError)
		return
	}
	
	workoutId, err := strconv.ParseInt(paramsWorkoutId, 10, 64) // Base 10, 64 bit int
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to parse integer ID", http.StatusInternalServerError)
		return
	}

	existingWorkout, err := wh.workoutStore.GetWorkoutById(workoutId)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to fetch workout", http.StatusInternalServerError)
		http.NotFound(w, r)
		return
	}

	if existingWorkout == nil {
		fmt.Println("Failed to fetch workout")
		http.Error(w, "Failed to fetch workout", http.StatusInternalServerError)
		return
	}

	var updateWorkoutRequest store.UpdateWorkout

	err = json.NewDecoder(r.Body).Decode(&updateWorkoutRequest)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to decode JSON", http.StatusInternalServerError)
		return
	}

	if updateWorkoutRequest.Title != nil {
		existingWorkout.Title = *updateWorkoutRequest.Title // Assign the value of the pointer to the existing workout, otherwise leave as is
	}

	if updateWorkoutRequest.Description != nil {
		existingWorkout.Description = *updateWorkoutRequest.Description
	}

	if updateWorkoutRequest.DurationMinutes != nil {
		existingWorkout.DurationMinutes = *updateWorkoutRequest.DurationMinutes
	}

	if updateWorkoutRequest.CaloriesBurned != nil {
		existingWorkout.CaloriesBurned = *updateWorkoutRequest.CaloriesBurned
	}

	if updateWorkoutRequest.Entries != nil {
		existingWorkout.Entries = updateWorkoutRequest.Entries // entries zero val is nil so doesn't need a pointer
	}

	err = wh.workoutStore.UpdateWorkout(existingWorkout, workoutId)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to update workout", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(existingWorkout)
}

func (wh *WorkoutHandler) HandleDeleteWorkout(w http.ResponseWriter, r *http.Request) {
	paramsWorkoutId := chi.URLParam(r, "id")
	if paramsWorkoutId == "" {
		fmt.Println("No ID provided")
		http.Error(w, "No ID provided", http.StatusInternalServerError)
		return
	}
	
	workoutId, err := strconv.ParseInt(paramsWorkoutId, 10, 64) // Base 10, 64 bit int
	if err != nil {
		fmt.Println(err)
		http.Error(w, "failed to parse integer ID", http.StatusInternalServerError)
		return
	}

	err = wh.workoutStore.DeleteWorkout(workoutId)
	if err == sql.ErrNoRows {
		http.Error(w, "workout not found", http.StatusNotFound)
		return
	}
	if err != nil {
		fmt.Println(err)
		http.Error(w, "failed to delete workout", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}