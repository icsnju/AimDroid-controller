# README


This project hosts the implementation of the controller of AimDroid.
The internal name of the controller is `monidroid`

## Build

1. Install `golang`.
    1. Mac OS: `brew install go`
2. Create the workspace for `go` following <https://golang.org/doc/code.html>. The following is a list of simplified steps.
    1. Create the default go workspace at `$HOME/go/`.
    2. Setup env variable `GOPATH`: `export GOPATH=$HOME/go`
    3. Create the default go source code folder at `$HOME/go/src`.
3. Change directory to `$HOME/go/src` and clone the source code of Monidroid: `git clone `
4. Change directory to `$HOME/go/src/Monidroid` and run `go build`.
5. You will see an executable tool named `monidroid` at `$HOME/go/src/Monidroid/`


## Configuration


Here is a sample of configuration file.

```
{
    "PackageName":"com.cvicse.zhnt",
    "MainActivity":"com.cvicse.zhnt.LoadingActivity",
    "SDKPath":"/home/dmemo/Android/android-sdk-linux",
    "Epsilon":0.2,
    "Alpha":0.6,
    "Gamma":0.5,
    "MaxSeqLen":100,
    "MinSeqLen":20,
    "Time":3600
}
```

Here is the explanation of each option.
1. `PackageName`: the app name of the app under test.
2. `MainActivity`: the entry activity of the app.
3. `SDKPath`: the path to the Android SDK
    * We need to run the command `adb` to communicate with the phone.
4. `Epsilon`, `Alpha`, `Gamma`: the parameters for the reinforcement learning module. See the paper for more details.
5. `MaxSeqLen`, `MinSeqLen`: the length of actions in a single activity. See the paper for more details.
6. `Time`: the total testing time in seconds.
