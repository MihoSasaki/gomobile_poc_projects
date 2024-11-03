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

## imageload
Sample application of loading texture by png image.
This is to show how to load png image and apply texture to the application

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