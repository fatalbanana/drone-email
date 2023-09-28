# drone-email

[![Go Report](https://goreportcard.com/badge/github.com/fatalbanana/drone-email)](https://goreportcard.com/report/github.com/fatalbanana/drone-email)

Drone plugin to send build status notifications via Email. For the usage information and a listing of the available options please take a look at [the docs](DOCS.md).

### Example
Execute from the working directory:

```sh
docker run --rm \
  -e PLUGIN_FROM.ADDRESS=drone@test.test \
  -e PLUGIN_FROM.NAME="John Smith" \
  -e PLUGIN_HOST=smtp.test.test \
  -e PLUGIN_USERNAME=drone \
  -e PLUGIN_PASSWORD=test \
  -e DRONE_REPO_OWNER=octocat \
  -e DRONE_REPO_NAME=hello-world \
  -e DRONE_COMMIT_SHA=7fd1a60b01f91b314f59955a4e4d4e80d8edf11d \
  -e DRONE_COMMIT_BRANCH=master \
  -e DRONE_COMMIT_AUTHOR=octocat \
  -e DRONE_COMMIT_AUTHOR_EMAIL=octocat@test.test \
  -e DRONE_BUILD_NUMBER=1 \
  -e DRONE_BUILD_STATUS=success \
  -e DRONE_BUILD_LINK=http://github.com/octocat/hello-world \
  -e DRONE_COMMIT_MESSAGE="Hello world!" \
  -v $(pwd):$(pwd) \
  -w $(pwd) \
  nerfd/drone-email
```
