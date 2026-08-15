Currently we have two functions that start a docker container in `pkg/docker/client.go`.

We only want one unified function that satifies all requirements.
Drop the simple version and use the complex version always.
Make sure proper defaults are applied when options are not actually used.