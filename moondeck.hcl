klipper {
  api = "printer-delta.local:7125"
}

menu "home" {
  icon = "bars"
  item = [
    {
      app  = "home"
      icon = "home"
    },
    {
      app  = "move"
      icon = "crosshairs"
    },
    {
      app  = "camera"
      icon = "camera"
    }
  ]
  fg = "#ffcc80"
  bg = "#424242"
}

preset = [
  { name = "PLA", extruder = 210, bed = 60 },
  { name = "ABS", extruder = 250, bed = 90 },
]

app "home" {
  background = "klipper.png"

  emergency {
    at = {x: 0, y: 2}
    confirm = false
  }

  menu "home" {
    at = {x: 0, y: 0}
  }

  gcode "motors-off" {
    at = {x: 0, y: 1}
    icon = "motor-off"
    gcode = <<END
  M84
  M117 Motors off
  END
  }

  icon "extruder" {
    at = {x: 1, y: 0}
  }

  temp "extruder" {
    at = {x: 2, y: 0}
    graph = {w: 2, h: 1}
    fg = "#ffcc80"
    bg = "#000000"
  }

  icon "bed" {
    at = {x: 1, y: 1}
  }

  temp "heater_bed" {
    at = {x: 2, y: 1}
    graph = {w: 2, h: 1}
    fg = "#ffcc80"
    bg = "#000000"
  }
}

app "move" {
  background = "klipper.png"

  emergency {
    at = {x: 0, y: 2}
    confirm = false
  }

  menu "home" {
    at = {x: 0, y: 0}
  }

  move {
    at = {x: 1, y: 0}
  }
}

app "camera" {
   emergency {
    at = {x: 0, y: 2}
    confirm = false
  }

  menu "home" {
    at = {x: 0, y: 0}
  }

  camera {
    at  = {x: 1, y: 0}
    url = "http://printer-delta.local/webcam/?action=stream"
    fps = 5
  }
}