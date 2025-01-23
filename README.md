# Game Forum

This project consists in creating a web forum that allows :
<ul>
    <li>communication between users.</li>
    <li>associating categories to posts.</li>
    <li>liking and disliking posts and comments.</li>
    <li>filtering posts.</li>
</ul>

Implemented: advanced-features, authentication(google, github), image-upload, moderation, security (created our own certificate)

## Contributors
<ol>
    <li>Abay Aliyev</li>
    <li>Adilzhan Shirbayev</li>
</ol>

## Dependencies
<ul>
    <li>Go programming language</li>
    <li>Sqlite</li>
    <li><a href="golang.org/x/crypto">crypto</a></li>
    <li>Selenium for testing</li>
    <li>UUID</li>
    <li>Docker</li>
</ul>

## How to run
By default, our project runs on :8433 port
### Using Docker and makefile
run ```make all``` if you running first time. <br>
Alternative: ```make build``` to build, then ```make run``` to run, and open https://localhost:8433/ 

### Using go run
run ```go run ./cmd/web``` and open https://localhost:8433/ 

