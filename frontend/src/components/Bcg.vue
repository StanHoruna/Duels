<template>
  <div class="bcg">
    <canvas id='backgroundCanvas'></canvas>
    <div class="bgs" id="nightBg"></div>
  </div>
</template>

<script setup>
import {onMounted} from "vue";

class Canvas {

  constructor() {
    this.sprites = []
  }

  new(id, width, height) {
    this.ctx = document.getElementById(id).getContext('2d')
    this.ctx.canvas.width = width
    this.ctx.canvas.height = height
  }

  resize(width, height) {
    this.ctx.canvas.width = width
    this.ctx.canvas.height = height
  }

  line(startX, startY, endX, endY, strokeColor, lineWidth, lineCap) {
    this.ctx.beginPath()
    this.ctx.moveTo(startX, startY)
    this.ctx.lineTo(endX, endY)
    if (lineCap) this.ctx.lineCap = lineCap
    this.paint(lineWidth || 1, strokeColor || 'black')
    this.ctx.fill()
    this.ctx.stroke()
  }

  text(content, x, y, font, style, fillColor) {
    this.ctx.font = font + 'px ' + style
    this.ctx.fillStyle = fillColor
    this.ctx.fillText(content, x, y)
  }

  paint(width, stroke, fill) {
    this.ctx.fillStyle = fill
    this.ctx.lineWidth = width
    this.ctx.strokeStyle = stroke
  }

  lGrad(obj, g = obj.gradient, cols = obj.colors) {
    var gradient = this.ctx.createLinearGradient(g[0], g[1], g[2], g[3])
    for (const c of cols) {
      gradient.addColorStop(c[0], c[1])
    }
    return gradient
  }

  clear() {
    this.ctx.clearRect(0, 0, this.ctx.canvas.width, this.ctx.canvas.height)
  }

  animate(end = true) {
    this.clear()
    this.sprites.map(x => {
      x.xPx *= x.r, x.yPx *= x.r
    })
    for (const s of this.sprites) {
      [s.xp, s.yp] = [s.xp + (s.xPx * s.s), s.yp + (s.yPx * s.s)]
      if ((s.yd === 1 && s.yp < s.ey) || (s.yd === -1 && s.yp > s.ey)
          || (s.xd === 1 && s.xp < s.ex) || (s.xd === -1 && s.xp > s.ex)) {
        this.line(s.x, s.y, s.xp, s.yp, s.c, s.stk)
        end = false
      } else {
        this.line(s.x, s.y, s.ex, s.ey, s.c, s.stk)
      }
    }
    if (!end) {
      window.requestAnimationFrame(() => {
        this.animate()
      })
    }
  }

}

class BackgroundEffect {

  constructor(params) {
    Object.assign(this, params)
    this.oldSpeed = this.speed
    this.cx = window.innerWidth / 2
    this.cy = window.innerHeight / 2
    this.longer = this.cx > this.cy ? this.cx : this.cy
    this.shorter = this.cx > this.cy ? this.cy : this.cx
    this.lines = new Array(this.linesCnt).fill({})
    this.c = new Canvas
  }

  startBackground() {
    this.c.new('backgroundCanvas', window.innerWidth, window.innerHeight)
    this.lines = this.lines.map(l => this.newAll())
    this.setSpeed(this.speed)
    this.move() // setTimeout(()=>{ move() },1000)
  }

  move() {
    this.c.clear()
    this.c.resize(window.innerWidth, window.innerHeight)
    if (this.speed !== this.oldSpped) {
      this.lines.map(l => {
        l.speed = _rand(3, 10) * this.speed
      })
      this.oldSpped = this.speed
    }
    if (this.effect === "lightSpeed") this.moveLightSpeed()
    if (this.effect === "racers") this.moveRacers()
    if (!this.pause) window.requestAnimationFrame(() => {
      this.move()
    })
  }

  moveRacers() {
    this.lines.forEach((l, i) => {
      // 🗺 Check for new locaiton
      if (l.x > this.lineW + window.innerWidth || isNaN(l.x)) {
        this.lines[i] = l = this.newRacers()
      }
      l.x += l.speed
      let color = this.c.lGrad({
        gradient: [l.x, 0, l.x + (this.lineW * this.speed), 0],
        colors: l.colors
      })
      this.c.line(l.x, l.y, l.x + (this.lineW * this.speed), l.y, color, l.lineS, "round")
    })
  }

  moveLightSpeed() {
    this.lines.forEach((l, i) => {
      // 🗺 Check for new locaiton
      if (l.dist - this.lineW > this.longer / 2) {
        this.lines[i] = l = this.newLightSpeed()
      }
      l.dist += l.speed
      l.x = (l.dist * Math.cos(l.angle)) + this.cx
      l.y = (l.dist * Math.sin(l.angle)) + this.cy
      const l2x = ((l.dist + (this.lineW * this.speed)) * Math.cos(l.angle)) + this.cx
      const l2y = ((l.dist + (this.lineW * this.speed)) * Math.sin(l.angle)) + this.cy
      let color = this.c.lGrad({
        gradient: [l.x, l.y, l2x, l2y],
        colors: l.colors
      })
      this.c.line(l.x, l.y, l2x, l2y, color, l.lineS, "round")
    })
  }

  newRacers() {
    const {dist, degree, angle, lineS, colors} = this.angles()
    return {
      dist, degree, angle, lineS, colors,
      x: _rand(-1000, this.lineW * -1),
      y: (_rand(0, ((window.innerHeight - 5) / this.grid)) * this.grid - 16),
      speed: _rand(3, 10) * this.speed,
    }
  }

  newLightSpeed() {
    const {dist, degree, angle, lineS, colors} = this.angles()
    return {
      dist, degree, angle, lineS, colors,
      speed: _rand(3, 10) * this.speed,
      x: (dist * Math.cos(angle)) + this.cx,
      y: (dist * Math.sin(angle)) + this.cy,
    }
  }

  newAll() {
    const {dist, degree, angle, lineS, colors} = this.angles()
    return {
      dist, degree, angle, lineS, colors,
      speed: _rand(3, 10) * this.speed,
      x: window.innerWidth * 2,
      y: window.innerHeight * 2
    }
  }

  angles() {
    const dist = (this.longer) * (_rand(0, 20) / 10)
    const degree = _rand(0, 359)
    const angle = (degree * Math.PI / 180)
    const lineS = _rand(1, 6)
    let colors;
    if (this.setting === "night") {
      const [r, g, b] = [200 + _rand(-20, 20), 200 + _rand(-20, 20), 230 + _rand(-20, 20)]
      const o = _rand(2, 10) / 10
      const col = "rgba(" + r + "," + g + "," + b + ","
      colors = [["0.1", col + "0.0)"], ["1", col + o + ")"]]
    } else {
      const [r, g, b] = [100 + _rand(-30, 30), 149 + _rand(-30, 30), 237 + _rand(-20, 20)]
      const o = _rand(2, 10) / 10
      const col = "rgba(" + r + "," + g + "," + b + ","
      colors = [["0.1", col + "0.0)"], ["1", col + o + ")"]]
    }
    return {dist, degree, angle, lineS, colors}
  }

  speedUp(speed, max, ramp, down) {
    window.requestAnimationFrame(() => {
      if (!down && this.speed <= max) {
        this.setSpeed(this.speed *= ramp)
        this.speedUp(speed, max, ramp)
      } else if (down && this.speed >= max) {
        this.setSpeed(this.speed *= ramp)
        this.speedUp(speed, max, ramp, down)
      } else {
        this.setSpeed(max)
      }
    })
  }

  setSpeed(speed) {
    this.speed = speed
  }
}

let Bg;

function start() {
  Bg = new BackgroundEffect({
    effect: "lightSpeed", // lightSpeed, racers
    setting: "night",
    grid: 1, // (50 or 1) in pixed fixed position
    speed: 0.5, // Max = 1.0
    lineS: 3, // og: 0.4
    linesCnt: 150,
    lineW: 300,
    pause: false,
  })

  Bg.startBackground()

  setEvents()

  drawBox()
}

function setEvents() {
  window.onresize = function () {
    Bg.cx = window.innerWidth / 2
    Bg.cy = window.innerHeight / 2
    Bg.longer = Bg.cx > Bg.cy ? Bg.cx : Bg.cy
    Bg.shorter = Bg.cx > Bg.cy ? Bg.cy : Bg.cx
  }
}

function _rand(min, max) {
  return Math.floor(Math.random() * (max - min + 1) + min);
}

function drawBox() {
  const canvases = document.getElementsByClassName('draw-box')
  let c2 = []
  for (var i = 0; i < canvases.length; i++) {
    const w = canvases[i].parentElement.clientWidth + 8
    const h = canvases[i].parentElement.clientHeight + 9
    canvases[i].id = 'landing-canvas' + i
    c2[i] = new Canvas
    c2[i].new(canvases[i].id, w, h)
    c2[i].line(0, 5, w, 5, 'rgba(100,149,237,0.6)', 2, "round")
    c2[i].line(w, h - 4, 0, h - 4, 'rgba(100,149,237,0.6)', 2, "round")
    c2[i].line(4, 0, 4, h, 'rgba(100,149,237,0.6)', 2, "round")
    c2[i].line(w - 4, h, w - 4, 0, 'rgba(100,149,237,0.6)', 2, "round")
  }
}

onMounted(() => {
  start()
  Bg.speedUp(Bg.speed, 0.5, 1.04)
})
</script>

<style scoped lang="scss">
.bgs {
  position: absolute;
  top: 0; bottom: 0; left: 0; right: 0;
}
#nightBg {
  //background: url("https://s3-us-west-2.amazonaws.com/s.cdpn.io/782173/codepen-bg-dark1.png");
  background: rgba(0, 0, 0, 0.8);
  //background: #141414;
  z-index: -3;
}
.draw-box-container {
  position: relative;
}
.draw-box {
  position: absolute;
  top: -5px;
  left: -4px;
  pointer-events: none;
}
#backgroundCanvas {
  position: absolute;
  top: 0px; left: 0px;
  z-index: 0;
}

@media (max-width: 500px) {
  .bcg {
    display: none;
  }
}
</style>
