$VERSION = '0.1.0-winhack'
$GOPATH = (go env -json | ConvertFrom-Json).GOPATH
$BUILD = $GOPATH + '\src\k8s.io\kops\.build'
$DIST = $BUILD + "\dist"
New-Item -Force -ItemType Directory -Path $DIST
$GITSHA = Invoke-Command {cd "$GOPATH/src/k8s.io/kops"; git describe --always }
$env:GOOS='windows'
$Env:GOARCH='amd64'
go build -installsuffix cgo -o "$($DIST)/windows/amd64/kops.exe" -ldflags="-s -w -X k8s.io/kops.Version=$($VERSION) -X k8s.io/kops.GitVersion=$($GITSHA)" k8s.io/kops/cmd/kops