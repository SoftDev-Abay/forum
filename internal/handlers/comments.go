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

func (app *Application) createCommentPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		app.clientError(w, r, http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		app.clientError(w, r, http.StatusBadRequest)
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
		return
	}

	commentID, err := app.Comments.Insert(postId, userId, text, time.Now())
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// 1) Notify the post owner (except if the commenter is the same user).
	post, err := app.Posts.Get(postId)
	if err == nil && post.OwnerID != userId {
		// Insert a new notification of type "comment".
		// The actor is userId, the recipient is post.OwnerID,
		// the post is postId, and the new comment's ID is 'commentID'.
		_, nErr := app.Notifications.Insert(
			"comment",
			userId,           // actor
			post.OwnerID,     // recipient
			postId,           // post
			&commentID,       // comment
		)
		if nErr != nil {
			app.serverError(w, r, nErr)
			return
		}
	}

	// Redirect to the post view page
	http.Redirect(w, r, fmt.Sprintf("/post/view?id=%d", postId), http.StatusSeeOther)
}


func (app *Application) handleCommentReaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		app.clientError(w, r, http.StatusMethodNotAllowed)
		return
	}
	// Get the comment ID from the query parameters
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, r, http.StatusBadRequest)
		return
	}

	commentIDStr := r.URL.Query().Get("id")
	commentID, err := strconv.Atoi(commentIDStr)
	if err != nil || commentID < 1 {
		app.clientError(w, r, http.StatusBadRequest)
		return
	}

	postId, err := strconv.Atoi(r.PostForm.Get("postId"))
	if err != nil || postId < 1 {
		app.notFound(w, r)
		return
	}

	// Get the reaction type (like or dislike) from the form
	reaction := r.FormValue("reaction")
	if reaction != "like" && reaction != "dislike" {
		app.clientError(w, r, http.StatusBadRequest)
		return
	}
	// Get the user ID from the session or context
	userID, err := app.getAuthenticatedUserID(r)
	if err != nil {
		// If not authenticated, return an error or handle gracefully
		app.notAuthenticated(w, r)
		return
	}

	// Check current user reaction to decide if they are changing their reaction
	existingReaction, err := app.CommentsReactions.GetReaction(userID, commentID)
	if err != nil && err != models.ErrNoReaction {
		app.serverError(w, r, err)
		return
	}

	// Fetch the comment
	comment, err := app.Comments.Get(commentID)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	// Initialize like and dislike counts
	newLikeCount := comment.LikeCount
	newDislikeCount := comment.DislikeCount

	if existingReaction != nil {
		// User already reacted, handle toggling reactions
		if existingReaction.Type == reaction {
			// If they click on the same reaction, it will be removed (toggle)
			err = app.CommentsReactions.DeleteReaction(userID, commentID)
			if err != nil {
				app.serverError(w, r, err)
				return
			}

			if reaction == "like" {
				newLikeCount -= 1
			} else {
				newDislikeCount -= 1
			}

		} else {
			// If they switch reactions, update accordingly
			err = app.CommentsReactions.UpdateReaction(userID, commentID, reaction)
			if err != nil {
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
		}
	} else {
		// No existing reaction, so we add the new one
		err = app.CommentsReactions.AddReaction(userID, commentID, reaction)
		if err != nil {
			app.serverError(w, r, err)
			return
		}

		if reaction == "like" {
			newLikeCount += 1
		} else {
			newDislikeCount += 1
		}
	}

	// Update the like and dislike counts on the comment
	err = app.Comments.UpdateCommentLikeDislikeCounts(commentID, newLikeCount, newDislikeCount)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	if comment.UserID != userID {
		var notifType string
		if reaction == "like" {
			notifType = "comment_like"
		} else {
			notifType = "comment_dislike"
		}

		_, nErr := app.Notifications.Insert(
			notifType,
			userID,          // actor
			comment.UserID,  // recipient
			comment.PostID,  // post
			&comment.ID,     // comment
		)
		if nErr != nil {
			app.serverError(w, r, nErr)
			return
		}
	}

	// After updating, redirect to the comment view to update the UI
	redirectURL := fmt.Sprintf("/post/view?id=%d", postId)
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

func (app *Application) commentDelete(w http.ResponseWriter, r *http.Request) {
	// Only allow POST (or possibly DELETE, if you are using REST conventions).
	if r.Method != http.MethodPost {
		app.clientError(w, r, http.StatusMethodNotAllowed)
		return
	}

	// Parse the `id` from the query parameters (e.g. /post/delete?id=123).
	idStr := r.URL.Query().Get("id")
	commentID, err := strconv.Atoi(idStr)
	if err != nil || commentID < 1 {
		app.clientError(w, r, http.StatusBadRequest)
		return
	}

	// (Optional) Check if the user is authenticated and/or is allowed to delete the post.
	_, err = app.getAuthenticatedUserID(r)
	if err != nil {
		app.notAuthenticated(w, r)
		return
	}

	// Retrieve the post to get the image URL
	comment, err := app.Comments.Get(commentID)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	err = app.CommentsReactions.DeleteReactioByCommentId(commentID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	err = app.Comments.DeleteCommentById(commentID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// Redirect to the post view page
	http.Redirect(w, r, fmt.Sprintf("/post/view?id=%d", comment.PostID), http.StatusSeeOther)
}
