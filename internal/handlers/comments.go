package handlers

import (
	"errors"
	"fmt"
	"game-forum-abaliyev-ashirbay/internal/models"
	"net/http"
	"strconv"
	"time"
	// "game-forum-abaliyev-ashirbay/internal/models"
)


func (app *Application) createComment(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Comment create accessed")

	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	postId, err := strconv.Atoi(r.PostForm.Get("postId"))
	if err != nil || postId < 1 {
		app.notFound(w, r)
		return
	}

	text := r.PostForm.Get("text")

	userId, err := app.getAuthenticatedUserID(r)
	if err != nil {
		app.notAuthenticated(w, r)
	}

	_, err = app.Comments.Insert(postId, userId, text, time.Now())
	// commentId, err := app.Comments.Insert(postId, userId, text)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	fmt.Println("Comment created successfully")
}
func (app *Application) handleCommentReaction(w http.ResponseWriter, r *http.Request) {
	// Get the comment ID from the query parameters
	commentIDStr := r.URL.Query().Get("id")
	commentID, err := strconv.Atoi(commentIDStr)
	if err != nil || commentID < 1 {
		fmt.Println("Error: Invalid comment ID or comment ID less than 1")
		app.clientError(w, http.StatusBadRequest)
		return
	}
	fmt.Println("Comment ID:", commentID)

	// Get the reaction type (like or dislike) from the form
	reaction := r.FormValue("reaction")
	if reaction != "like" && reaction != "dislike" {
		fmt.Println("Error: Invalid reaction type:", reaction)
		app.clientError(w, http.StatusBadRequest)
		return
	}
	fmt.Println("Reaction type:", reaction)

	// Get the user ID from the session or context
	userID, err := app.getAuthenticatedUserID(r)
	if err != nil {
		// If not authenticated, return an error or handle gracefully
		fmt.Println("Error: User not authenticated")
		app.notAuthenticated(w, r)
		return
	}
	fmt.Println("User ID:", userID)

	// Check current user reaction to decide if they are changing their reaction
	existingReaction, err := app.CommentsReactions.GetReaction(userID, commentID)
	if err != nil && err != models.ErrNoReaction {
		fmt.Println("Error: Unable to get existing reaction:", err)
		app.serverError(w, r, err)
		return
	}
	fmt.Println("Existing reaction:", existingReaction)

	// Fetch the comment
	comment, err := app.Comments.Get(commentID)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			fmt.Println("Error: Comment not found")
			app.notFound(w, r)
		} else {
			fmt.Println("Error: Server error fetching comment:", err)
			app.serverError(w, r, err)
		}
		return
	}
	fmt.Println("Comment fetched:", comment)

	// Initialize like and dislike counts
	newLikeCount := comment.LikeCount
	newDislikeCount := comment.DislikeCount
	fmt.Println("Initial Like Count:", newLikeCount, "Initial Dislike Count:", newDislikeCount)

	if existingReaction != nil {
		// User already reacted, handle toggling reactions
		if existingReaction.Type == reaction {
			// If they click on the same reaction, it will be removed (toggle)
			fmt.Println("User clicked the same reaction, deleting reaction.")
			err = app.CommentsReactions.DeleteReaction(userID, commentID)
			if err != nil {
				fmt.Println("Error: Deleting reaction failed:", err)
				app.serverError(w, r, err)
				return
			}

			if reaction == "like" {
				newLikeCount -= 1
			} else {
				newDislikeCount -= 1
			}
			fmt.Println("Updated Like Count:", newLikeCount, "Updated Dislike Count:", newDislikeCount)

		} else {
			// If they switch reactions, update accordingly
			fmt.Println("User switched reaction, updating reaction.")
			err = app.CommentsReactions.UpdateReaction(userID, commentID, reaction)
			if err != nil {
				fmt.Println("Error: Updating reaction failed:", err)
				app.serverError(w, r, err)
				return
			}

			if reaction == "like" {
				newLikeCount += 1
				newDislikeCount -= 1
			} else {
				newLikeCount -= 1
				newDislikeCount += 1
			}
			fmt.Println("Updated Like Count:", newLikeCount, "Updated Dislike Count:", newDislikeCount)
		}
	} else {
		// No existing reaction, so we add the new one
		fmt.Println("No existing reaction, adding new one.")
		err = app.CommentsReactions.AddReaction(userID, commentID, reaction)
		if err != nil {
			fmt.Println("Error: Adding reaction failed:", err)
			app.serverError(w, r, err)
			return
		}

		if reaction == "like" {
			newLikeCount += 1
		} else {
			newDislikeCount += 1
		}
		fmt.Println("Updated Like Count:", newLikeCount, "Updated Dislike Count:", newDislikeCount)
	}

	// Update the like and dislike counts on the comment
	err = app.Comments.UpdateCommentLikeDislikeCounts(commentID, newLikeCount, newDislikeCount)
	if err != nil {
		fmt.Println("Error: Updating comment counts failed:", err)
		app.serverError(w, r, err)
		return
	}
	fmt.Println("Successfully updated comment like/dislike counts.")

	// After updating, redirect to the comment view to update the UI
	redirectURL := fmt.Sprintf("/comment/view?id=%d", commentID)
	fmt.Println("Redirecting to:", redirectURL)
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}
