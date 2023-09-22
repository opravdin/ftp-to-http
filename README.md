# FTP to HTTP
This service provides single time access to specific item on FTP server

## Why?
I am using this simple container to forward CCTV snapshots to my Telegram Bot. Telegram API requires you to send URL of picture which bot will send to chat. This is an ideal solution for such single time file access application. My script triggers the access to file and then passes URL to telegram API for 1-time download.

## Usage
1. Create access token for your specific file:
```
curl --request POST \
  --url http://localhost:2180/open \
  --header 'Content-Type: multipart/form-data' \
  --form url=ftp://user:password@ftpserver:port/folder1/folder2/item.jpg \
  --form key=somekey
```
Server will respond with:
```
{
  "token": "f265ef11-9046-495f-b740-66998eb8b46b"
}
```
2. Use this token to access file on your server

```
curl --request GET \
  --url 'http://localhost:2180/get?token=f265ef11-9046-495f-b740-66998eb8b46b'
```

Server will respond with the body of your file

## Limitations
This app keeps list of allowed files in-memory, which is completely wrong for redundancy-related tasks and for scaling.