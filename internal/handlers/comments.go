package handlers

import (
	"errors"
	"fmt"
	"game-forum-abaliyev-ashirbay/internal/models"
	"net/http"
	"strconv"
	"time"
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
	post, err := app.Posts.Get(postId)
	if err == nil && post.OwnerID != userId {
		_, nErr := app.Notifications.Insert(
			"comment",
			userId,       // actor
			post.OwnerID, // recipient
			postId,       // post
			&commentID,   // comment
		)
		if nErr != nil {
			app.serverError(w, r, nErr)
			return
		}
	}

	http.Redirect(w, r, fmt.Sprintf("/post/view?id=%d", postId), http.StatusSeeOther)
}

func (app *Application) handleCommentReaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		app.clientError(w, r, http.StatusMethodNotAllowed)
		return
	}
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

	reaction := r.FormValue("reaction")
	if reaction != "like" && reaction != "dislike" {
		app.clientError(w, r, http.StatusBadRequest)
		return
	}
	userID, err := app.getAuthenticatedUserID(r)
	if err != nil {
		app.notAuthenticated(w, r)
		return
	}

	existingReaction, err := app.CommentsReactions.GetReaction(userID, commentID)
	if err != nil && err != models.ErrNoReaction {
		app.serverError(w, r, err)
		return
	}

	comment, err := app.Comments.Get(commentID)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	newLikeCount := comment.LikeCount
	newDislikeCount := comment.DislikeCount

	if existingReaction != nil {
		if existingReaction.Type == reaction {
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


			var notifType string
			if reaction == "like" {
				notifType = "comment_like"
			} else {
				notifType = "comment_dislike"
			}
	
			_, nErr := app.Notifications.Insert(
				notifType,
				userID,         // actor
				comment.UserID, // recipient
				comment.PostID, // post
				&comment.ID,    // comment
			)
			if nErr != nil {
				app.serverError(w, r, nErr)
				return
			}

		}
	} else {
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

		var notifType string
		if reaction == "like" {
			notifType = "comment_like"
		} else {
			notifType = "comment_dislike"
		}

		_, nErr := app.Notifications.Insert(
			notifType,
			userID,         // actor
			comment.UserID, // recipient
			comment.PostID, // post
			&comment.ID,    // comment
		)
		if nErr != nil {
			app.serverError(w, r, nErr)
			return
		}
	}

	err = app.Comments.UpdateCommentLikeDislikeCounts(commentID, newLikeCount, newDislikeCount)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	// if comment.UserID != userID && existingReaction != nil && existingReaction.Type == reaction {

	// }

	redirectURL := fmt.Sprintf("/post/view?id=%d", postId)
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

func (app *Application) commentDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		app.clientError(w, r, http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	commentID, err := strconv.Atoi(idStr)
	if err != nil || commentID < 1 {
		app.clientError(w, r, http.StatusBadRequest)
		return
	}

	userID, err := app.getAuthenticatedUserID(r)
	if err != nil {
		app.notAuthenticated(w, r)
		return
	}

	comment, err := app.Comments.Get(commentID)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	user, err := app.Users.GetById(userID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	if comment.UserID != userID && user.Role != "moderator" && user.Role != "admin" {
		app.clientError(w, r, http.StatusForbidden)
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

	http.Redirect(w, r, fmt.Sprintf("/post/view?id=%d", comment.PostID), http.StatusSeeOther)
}
