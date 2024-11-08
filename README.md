# gomobile PoC Project
This project is for demos applications build by gomobile.

## minigame
Sample application for 2D game by using [sprite engine](https://pkg.go.dev/golang.org/x/mobile/exp/sprite).

- The rocket moves up and down with touch or key events.
- Stars randomly appear from the right.
- Touching the area to the right of the rocket or pressing the space bar will release a bomb.
- If a star collides with the rocket, it will animate flying downward.
- If a collision occurs, the player loses a life.
- The game ends when the player's lives reach zero.

https://github.com/user-attachments/assets/7bd3d9e9-9eaf-47ff-9ebe-b5a2e8b59e2f


### How to run applications

Running these commands to execute.
```
$ go mod tidy
$ cd ./minigame
$ go run main.go
```

## imageload
Sample application of loading texture by png image.
This is to show how to load png image and apply texture to the application

<img width="470" alt="Screenshot 2024-11-05 at 22 26 53" src="https://github.com/user-attachments/assets/22c111f9-bffa-4b88-9141-657dbf0013c6">

### How to run applications

Running these commands to execute.
```
$ go mod tidy
$ cd ./imageload
$ go run main.go
```

## Current issues(I found out)
### Android
- All the loaded texture positions are not applied to Android application(even if I specify the x,y, it sets it to 0 position)
- minSDK issue(https://github.com/golang/mobile/pull/99/files)
  - Some newer devices may not download the application
### iOS
- Some version of iOS Simulator device cannot be used.(Crashed when opening)
- Code signing issue
  - Apple application needs code siging with the registed team. And in gomobile it is not automatically signed.
  - workaround(https://github.com/golang/go/issues/26615#issuecomment-451920252)

## Resources
minigame
- https://pngtree.com/freepng/space-game-asset-8bit-pixel-art-galaxy-planets_8673346.html
