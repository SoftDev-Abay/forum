userInfoJSON, err := json.Marshal(userInfo)
if err != nil {
    fmt.Println("Error marshaling user info:", err)
    return err
}

// Print the marshaled JSON for debugging
fmt.Println("User Info JSON:", string(userInfoJSON))

2. Ensure Valid Cookie Value

The cookie value should be a valid string. Ensure that the userInfoJSON does not contain any characters that are invalid for cookie values. If your user information is somehow causing issues (e.g., certain characters not being properly escaped), you can consider base64 encoding the JSON:

go

import "encoding/base64"

// Encode the JSON string
userInfoEncoded := base64.StdEncoding.EncodeToString(userInfoJSON)

userInfoCookie := http.Cookie{
    Name:     "user_info",
    Value:    userInfoEncoded,
    Path:     "/",
    MaxAge:   86400, // 1 day
    HttpOnly: true,
    Secure:   true,
    SameSite: http.SameSiteLaxMode,
}

3. Set the Cookie Properly

When setting cookies, make sure you're using http.SetCookie(w, &userInfoCookie) instead of r.AddCookie(&userInfoCookie). The SetCookie function is the correct way to add a cookie to the response:

go

http.SetCookie(w, &userInfoCookie)

4. Adjust Cookie Attributes

You might also want to review the attributes of your cookie. For instance, the MaxAge you've set seems to be quite high. You probably want a simpler representation:

go

MaxAge: 86400, // 1 day

5. Middleware for User Info

If you're concerned about the user info cookie, and if it’s acceptable within your application’s architecture, you can stick to storing only the token in the cookie. You can then retrieve user info from your database using the user ID associated with that token in your middleware. This approach often aligns with best practices since it reduces cookie size and potential security issues.
Example Middleware Approach

You can implement middleware like this:

go

func (app *Application) AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        tokenCookie, err := r.Cookie("token")
        if err != nil {
            http.Redirect(w, r, "/login", http.StatusSeeOther)
            return
        }

        // Fetch user info from the database using the token
        userInfo, err := app.Session.GetUserInfoByToken(tokenCookie.Value)
        if err != nil {
            http.Redirect(w, r, "/login", http.StatusSeeOther)
            return
        }

        // Attach userInfo to the request context
        ctx := context.WithValue(r.Context(), userContextKey, userInfo)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

This way, you only store the token in the cookie, and retrieve user info from your sessions table as needed.
Summary

    Ensure JSON is valid and properly formatted for cookies.
    Use http.SetCookie to set cookies correctly.
    Consider middleware to handle user info retrieval based on the token, reducing the need to store sensitive data in cookies.

Implement these changes, and your authentication mechanism should work smoothly!
Don't share sensitive info. Chats may be reviewed and used to train our models. Learn more


ChatGPT can make mistakes. Check important info.
