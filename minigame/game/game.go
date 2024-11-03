package game

import (
	"fmt"
	"image"
	"log"
	"math"
	"math/rand"

	_ "image/png"

	"github.com/google/uuid"
	"golang.org/x/mobile/asset"
	"golang.org/x/mobile/exp/f32"
	"golang.org/x/mobile/exp/sprite"
	"golang.org/x/mobile/exp/sprite/clock"
)

const (
	stoneWidth, stoneHeight = 20, 20 // width and height of each tile
	bombWidth, bombHeight   = 20, 20 // width and height of each tile
	tilesX, tilesY          = 16, 16 // number of horizontal tiles

	gameLaunchTime      = 5 * 60 // for about 5 sec, enemy is not coming
	probabilityOfEnemy  = 3      // how many enemies are coming
	cyclePeriod         = 15 * 60
	lostTimeBeforeReset = 240 // how long to wait before restarting the game

	rocketWidth, rocketHeight = 40, 40

	randomStoneNum = 4
	padding        = 10

	// heart information
	totalHp     = 5
	heartWidth  = 25
	heartHeight = 20
)

type Game struct {
	eng    sprite.Engine
	scene  *sprite.Node
	rocket struct {
		y float32 // y-offset
		v float32 // velocity
	}

	lastCalc        clock.Time // when we last calculated a frame
	nextCyclePeriod int32

	bombs     map[string]*Bomb
	bombNodes map[string]*sprite.Node

	crashedNodes map[string]*sprite.Node
	crashs       map[string]*Crash

	stones     map[string]*Stone
	stoneNodes map[string]*sprite.Node

	nextStones map[int]int

	width  float32
	height float32

	texs []sprite.SubTex

	currentHp int

	lost     bool
	lostTime clock.Time
}

type Stone struct {
	x   float32
	v   float32
	y   float32
	hit bool
}

type Bomb struct {
	x      float32
	v      float32
	y      float32
	hit    bool
	hitted bool
}

type Crash struct {
	crashTime clock.Time
}

func NewGame(e sprite.Engine, w, h float32) *Game {
	s := sprite.Node{}
	g := Game{
		eng:    e,
		scene:  &s,
		width:  w,
		height: h,
		lost:   false,
	}
	g.reset(0)
	return &g
}

func (g *Game) reset(now clock.Time) {
	g.rocket.y = g.height / 2
	g.rocket.v = 0

	g.stones = map[string]*Stone{}
	g.stoneNodes = map[string]*sprite.Node{}
	g.nextStones = map[int]int{}

	g.bombs = map[string]*Bomb{}
	g.bombNodes = map[string]*sprite.Node{}
	g.crashedNodes = map[string]*sprite.Node{}
	g.crashs = map[string]*Crash{}
	g.nextCyclePeriod = gameLaunchTime + int32(now)

	g.currentHp = totalHp
	g.lost = false
}

type arrangerFunc func(e sprite.Engine, n *sprite.Node, t clock.Time)

func (a arrangerFunc) Arrange(e sprite.Engine, n *sprite.Node, t clock.Time) { a(e, n, t) }

func newNode(fn arrangerFunc, eng sprite.Engine, scene *sprite.Node) *sprite.Node {
	n := &sprite.Node{Arranger: arrangerFunc(fn)}
	eng.Register(n)
	scene.AppendChild(n)

	return n
}

func removeNode(eng sprite.Engine, scene *sprite.Node, n *sprite.Node) {
	eng.Unregister(n)
	scene.RemoveChild(n)
}

func (g *Game) loadHPWithPosition(texs []sprite.SubTex, xPos, position int) {
	newNode(func(eng sprite.Engine, n *sprite.Node, t clock.Time) {
		a := f32.Affine{
			{heartWidth, 0, float32(xPos)},
			{0, heartHeight, padding},
		}

		if position < g.currentHp {
			eng.SetSubTex(n, texs[texHeartFull])
		} else {
			eng.SetSubTex(n, texs[texHeartEmpty])
		}

		eng.SetTransform(n, a)
	}, g.eng, g.scene)
}

func (g *Game) Scene() *sprite.Node {
	texs, err := loadTextures(g.eng)
	if err != nil {
		log.Fatalln(err)
	}
	g.texs = texs

	g.eng.Register(g.scene)
	g.eng.SetTransform(g.scene, f32.Affine{
		{1, 0, 0},
		{0, 1, 0},
	})

	heartXPos := padding
	for i := 0; i < totalHp; i++ {
		g.loadHPWithPosition(texs, heartXPos, i)
		heartXPos += heartWidth
	}

	g.addNewStar(g.width/3, g.height/3)
	g.addNewStar(g.width/3*2, g.height/3*2)

	// The rocket.
	newNode(func(eng sprite.Engine, n *sprite.Node, t clock.Time) {
		a := f32.Affine{
			{rocketWidth, 0, padding},
			// {0, rocketHeight, g.rocket.y - stoneHeight + stoneHeight/4},
			{0, rocketHeight, g.rocket.y - rocketHeight/2},
		}
		eng.SetSubTex(n, g.texs[texFlight])
		eng.SetTransform(n, a)
	}, g.eng, g.scene)

	return g.scene
}

func loadTextures(eng sprite.Engine) ([]sprite.SubTex, error) {
	a, err := asset.Open("minigame.png")
	if err != nil {
		return []sprite.SubTex{}, err
	}
	defer a.Close()

	m, _, err := image.Decode(a)
	if err != nil {
		return []sprite.SubTex{}, err
	}
	t, err := eng.LoadTexture(m)
	if err != nil {
		return []sprite.SubTex{}, err
	}

	return []sprite.SubTex{
		texFlight:          sprite.SubTex{t, image.Rect(0, 0, 64, 80)},
		texBeem1:           sprite.SubTex{t, image.Rect(70, 50, 126, 80)},
		texBeem2:           sprite.SubTex{t, image.Rect(128, 0, 173, 30)},
		texPlanet1:         sprite.SubTex{t, image.Rect(177, 0, 251, 83)},
		texPlanet2:         sprite.SubTex{t, image.Rect(251, 0, 325, 83)},
		texStarPinkFade1:   sprite.SubTex{t, image.Rect(325, 50, 388, 80)},
		texStarPinkBright1: sprite.SubTex{t, image.Rect(389, 50, 420, 80)},
		texStarBlueFade1:   sprite.SubTex{t, image.Rect(420, 0, 437, 37)},
		texStarPinkFade2:   sprite.SubTex{t, image.Rect(420, 0, 437, 37)},
		texStarPinkBright2: sprite.SubTex{t, image.Rect(420, 0, 437, 37)},
		texStarBlueBright1: sprite.SubTex{t, image.Rect(420, 0, 496, 37)},
		texCrash:           sprite.SubTex{t, image.Rect(496, 0, 560, 80)},
		texHeartFull:       sprite.SubTex{t, image.Rect(560, 60, 586, 80)},
		texHeartEmpty:      sprite.SubTex{t, image.Rect(560, 40, 586, 60)},
	}, nil
}

// down is checking whether press is released or not
func (g *Game) Press(keyPress bool, condition TouchCondition, isCursor bool) {
	if keyPress {
		if !isCursor {
			g.PopNewBomb()
			return
		}

		// when pressed on the above
		if condition == TouchUp && g.rocket.y > 0 {
			g.rocket.v -= 2.0
		}

		// when pressed on the below
		if condition == TouchDown && g.rocket.y < g.width {
			g.rocket.v += 2.0
		}
	} else {
		g.rocket.v = 0
	}
}

func (g *Game) CheckTouchIsCursor(touchX int) bool {
	return touchX < rocketWidth*3
}

func (g *Game) CheckTouchIsUp(touchY int) TouchCondition {
	yposition := touchY + int(g.height/2)
	if g.rocket.y >= float32(yposition) {
		return TouchUp
	} else if g.rocket.y < float32(yposition) {
		return TouchDown
	}

	return TouchStay
}

func (g *Game) PopNewBomb() {
	if g.lost {
		return
	}

	key := uuid.NewString()
	g.addNewBomb(key, &Bomb{
		rocketWidth+5,
		2,
		g.rocket.y,
		false,
		false,
	})
}

func (g *Game) calcStone() {
	for k, s := range g.stones {
		if !g.stones[k].hit {
			s.x = s.x - 1
			g.stones[k] = s
		}
		if s.x < stoneWidth*2 {
			g.currentHp -= 1
			if g.currentHp < 1 {
				g.lost = true
			}
		}

		if s.x < stoneWidth*2 || s.x > g.width || s.y < 0 || s.y > g.height {
			g.removeStone(k)
		}
	}
}

func (g *Game) calcBomb() {
	for k, s := range g.bombs {
		if g.bombs[k].hit {
			g.removeBomb(k)
			continue
		} else {
			s.x = s.x + 1
			g.bombs[k] = s
		}

		// remove crash and bomb
		if s.x < 0 || s.x > g.width || s.y < 0 || s.y > g.height {
			g.removeBomb(k)
		}
	}
}

func (g *Game) checkBombHits() {
	for bk, b := range g.bombs {
		for sk, s := range g.stones {
			if b.x+bombWidth >= s.x && b.x+bombWidth*2 > s.x {
				if b.y >= s.y-5 && b.y <= s.y+stoneHeight {
					fmt.Println("hit to the stone!")
					g.bombs[bk].hit = true
					g.stones[sk].hit = true
				}
			}
		}
	}
}

func (g *Game) checkCrashs(now clock.Time) {
	for k, c := range g.crashs {
		if (now - c.crashTime) > 15 {
			g.removeCrash(k)
		}
	}
}

func (g *Game) addNewStone(k string, s *Stone) {
	g.stones[k] = s
	n := newNode(func(eng sprite.Engine, n *sprite.Node, t clock.Time) {
		a := f32.Affine{
			{stoneWidth, 0, g.stones[k].x},
			{0, stoneHeight, g.stones[k].y},
		}
		if g.stones[k].hit {
			g.animateStone(&a, t, k)
		}

		eng.SetSubTex(n, g.texs[texPlanet1])
		eng.SetTransform(n, a)
	}, g.eng, g.scene)
	g.stoneNodes[k] = n
}

func (g *Game) addNewCrash(k string, x, y float32, ct clock.Time) {
	n := newNode(func(eng sprite.Engine, n *sprite.Node, t clock.Time) {
		ba := f32.Affine{
			{10, 0, x},
			{0, 10, y - 5},
		}
		ba.Scale(&ba, 1.5, 1.5)
		eng.SetSubTex(n, g.texs[texCrash])
		eng.SetTransform(n, ba)
	}, g.eng, g.scene)
	g.crashs[k] = &Crash{
		crashTime: ct,
	}
	g.crashedNodes[k] = n
}

func (g *Game) addNewBomb(k string, b *Bomb) {
	g.bombs[k] = b
	n := newNode(func(eng sprite.Engine, n *sprite.Node, t clock.Time) {
		a := f32.Affine{
			{20, 0, g.bombs[k].x},
			{0, 15, g.bombs[k].y},
		}

		bomb := g.bombs[k]
		if bomb.hit {
			if !bomb.hitted {
				g.addNewCrash(k, bomb.x-5, bomb.y+18, t)
			}
			g.bombs[k].hitted = true
		}

		eng.SetSubTex(n, g.texs[texBeem1])
		eng.SetTransform(n, a)
	}, g.eng, g.scene)
	g.bombNodes[k] = n
}

func (g *Game) addNewStar(x, y float32) {
	newNode(func(eng sprite.Engine, n *sprite.Node, t clock.Time) {
		a := f32.Affine{
			{20, 0, x},
			{0, 20, y},
		}

		x := frame(t, 20, texStarPinkBright1, texStarPinkFade1)
		eng.SetSubTex(n, g.texs[x])
		eng.SetTransform(n, a)
	}, g.eng, g.scene)
}

func (g *Game) animateStone(a *f32.Affine, t clock.Time, sk string) {
	dt := float32(t)
	a.Translate(a, 0.5, 0.5)
	a.Rotate(a, dt/math.Pi/-8)
	a.Translate(a, -0.5, -0.5)
	g.stones[sk].x = g.stones[sk].x + 0.5
	g.stones[sk].y = g.stones[sk].y + 1.5
}

func (g *Game) removeStone(key string) {
	n := g.stoneNodes[key]
	g.scene.RemoveChild(n)
	delete(g.stoneNodes, key)
	delete(g.stones, key)
}

func (g *Game) removeBomb(key string) {
	n := g.bombNodes[key]
	g.scene.RemoveChild(n)
	delete(g.bombNodes, key)
	delete(g.bombs, key)
}

func (g *Game) removeCrash(key string) {
	n := g.crashedNodes[key]
	g.scene.RemoveChild(n)
	delete(g.crashedNodes, key)
	delete(g.crashs, key)
}

func (g *Game) removeAll() {
	for k := range g.bombs {
		g.removeBomb(k)
	}

	for k := range g.stones {
		g.removeStone(k)
	}

	for k := range g.crashs {
		g.removeCrash(k)
	}
}

func (g *Game) popRandomStone(now clock.Time) {
	if now < gameLaunchTime {
		return
	}

	// next cycle is coming
	if now > clock.Time(g.nextCyclePeriod) {
		fmt.Println("next cycle coming!")
		g.nextCyclePeriod += cyclePeriod
		g.nextStones = g.pickupNums(int(now), int(g.nextCyclePeriod), probabilityOfEnemy)
	}

	_, exists := g.nextStones[int(now)]
	if exists {
		fmt.Println("new stone poped")
		key := uuid.NewString()
		stoneY := rand.Intn(int(g.height))
		g.addNewStone(
			key,
			&Stone{
				x:   g.width,
				y:   float32(stoneY),
				v:   2,
				hit: false,
			})
	}
}

func (g *Game) pickupNums(start, end, nums int) map[int]int {
	numRange := end - start

	selected := make(map[int]int)
	for counter := 0; counter < nums; {
		n := rand.Intn(numRange) + start
		_, exist := selected[n]
		if !exist {
			selected[n] = n
			counter++
		}
	}

	return selected
}

func (g *Game) calRocket() {
	// Compute offset.
	g.rocket.y += g.rocket.v
}

type TouchCondition int

const (
	TouchUp TouchCondition = iota
	TouchStay
	TouchDown
)

const (
	texFlight = iota
	texBeem1
	texBeem2
	texPlanet1
	texPlanet2
	texStarPinkFade1
	texStarPinkBright1
	texStarPinkBright2
	texStarBlueBright1
	texStarPinkFade2
	texStarBlueFade1
	texCrash
	texHeartFull
	texHeartEmpty
)

// frame returns the frame for the given time t
// when each frame is displayed for duration d.
func frame(t, d clock.Time, frames ...int) int {
	total := int(d) * len(frames)
	return frames[(int(t)%total)/int(d)]
}

func (g *Game) Update(now clock.Time) {
	if g.lost && now-g.lostTime > lostTimeBeforeReset {
		g.removeAll()
		g.reset(now)
	}

	// Compute game states up to now.
	for ; g.lastCalc < now; g.lastCalc++ {
		g.calcFrame(now)
	}
}

func (g *Game) calcFrame(now clock.Time) {
	g.calRocket()
	g.calcStone()
	g.calcBomb()
	g.checkBombHits()
	g.checkCrashs(now)
	g.popRandomStone(now)
}
