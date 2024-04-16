# LandslideVM Runner

Run following commands from [landslidevm](https://github.com/ConsiderItDone/landslidevm) repo to build subnet:

```shell
BASEDIR=/tmp/e2e-test-landslide AVALANCHEGO_BUILD_PATH=/tmp/e2e-test-landslide/avalanchego ./scripts/install_avalanchego_release.sh

./scripts/build.sh /tmp/e2e-test-landslide/avalanchego/plugins/pjSL9ksard4YE96omaiTkGL5H6XX2W5VEo3ZgWC9S2P6gzs9A
```

Then execute:

```shell
go run main.go
```