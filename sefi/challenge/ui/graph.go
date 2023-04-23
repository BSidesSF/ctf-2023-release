package main

import (
	"fmt"
	"log"

	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gtk"
)

type GraphRenderer interface {
	DrawFunc(*gtk.DrawingArea, *cairo.Context)
	SetData([]int)
	SetXRange(int, int)
	SetYRange(int, int)
}

type Color struct {
	red   float64
	green float64
	blue  float64
}

var (
	White = Color{red: 1.0, green: 1.0, blue: 1.0}
	Black = Color{}
	Blue  = Color{blue: 1.0}
	Red   = Color{red: 1.0}
)

// This is a very basic bargraph with no labels.
type BarGraphRenderer struct {
	data []int
	xmin int
	xmax int
	ymin int
	ymax int
	// Interval of major ticks
	xticksmajor int
	// Interval of minor ticks
	xticksminor     int
	fontsize        int
	bgcolor         Color
	fgcolor         Color
	barColorFunc    func(int, int) Color
	barDataProvider func(int) ([]int, float64)
}

func (r *BarGraphRenderer) DrawFunc(da *gtk.DrawingArea, cr *cairo.Context) {
	allocation := da.GetAllocation()
	width, height := allocation.GetWidth(), allocation.GetHeight()
	r.setDefaults(cr)
	r.drawBackground(cr, width, height)
	cr.SetSourceRGB(r.fgcolor.red, r.fgcolor.green, r.fgcolor.blue)
	tickHeight := r.drawTicks(cr, width, height)
	r.drawBars(cr, width, height-tickHeight)
}

func (r *BarGraphRenderer) SetData(data []int) {
	r.data = data
	r.xmin = 0
	r.xmax = len(data)
	r.ymax = 0
}

func (r *BarGraphRenderer) SetXRange(lo, high int) {
	r.xmin = lo
	r.xmax = high
}

func (r *BarGraphRenderer) SetYRange(lo, high int) {
	r.ymin = lo
	r.ymax = high
}

func (r *BarGraphRenderer) setDefaults(cr *cairo.Context) {
	// Update some things
	if r.fontsize == 0 {
		r.fontsize = 12
	}
	if r.xmax == 0 {
		r.xmax = len(r.data)
	}
	if r.ymax == 0 {
		for i := r.xmin; i < r.xmax && i < len(r.data); i++ {
			if r.data[i] > r.ymax {
				r.ymax = r.data[i]
			}
		}
	}
	if r.barColorFunc == nil {
		r.barColorFunc = r.defaultColorFunc
	}
	if r.barDataProvider == nil {
		r.barDataProvider = r.defaultDataProvider
	}
	// Set context defaults
	cr.SelectFontFace("Sans", cairo.FONT_SLANT_NORMAL, cairo.FONT_WEIGHT_NORMAL)
	cr.SetFontSize(float64(r.fontsize))
}

func (r *BarGraphRenderer) drawBackground(cr *cairo.Context, width, height int) {
	cr.ResetClip()
	cr.Rectangle(0, 0, float64(width), float64(height))
	cr.SetSourceRGB(r.bgcolor.red, r.bgcolor.green, r.bgcolor.blue)
	cr.Fill()
}

func (r *BarGraphRenderer) getTickLabelSize(cr *cairo.Context) (int, int) {
	extents := cr.TextExtents(fmt.Sprintf("%d", r.xmax))
	return int(extents.Width), int(extents.Height)
}

func (r *BarGraphRenderer) drawTicks(cr *cairo.Context, width, height int) int {
	tickHeight := 0
	labelMargin := 2
	majorTicklen := 8
	labelWidth, _ := r.getTickLabelSize(cr)
	if r.xticksmajor > 0 {
		numMajorTicks := (r.xmax - r.xmin) / r.xticksmajor
		for numMajorTicks*labelWidth > width {
			r.xticksmajor *= 2
			numMajorTicks = (r.xmax - r.xmin) / r.xticksmajor
		}
		cr.SetLineWidth(2.0)
		tickSpacing := float64(width*r.xticksmajor) / float64(r.xmax-r.xmin)
		halfBar := float64(width) / float64(r.xmax-r.xmin) / 2.0
		for i := 0; i <= numMajorTicks; i++ {
			lblText := fmt.Sprintf("%d", r.xmin+(i*r.xticksmajor))
			extents := cr.TextExtents(lblText)
			xpos := float64(i)*tickSpacing + halfBar
			ypos := height - labelMargin
			cr.NewPath()
			tickStart := float64(ypos-1) - extents.Height
			cr.MoveTo(xpos, tickStart)
			cr.LineTo(xpos, tickStart-float64(majorTicklen))
			cr.Stroke()
			xpos -= (extents.Width / 2)
			if xpos < 0.0 {
				xpos = 0.0
			}
			cr.MoveTo(xpos, float64(ypos))
			cr.ShowText(lblText)
			h := int(extents.Height) + labelMargin + majorTicklen + 1
			if h > tickHeight {
				tickHeight = h
			}
		}
	}
	if r.xticksminor > 0 {
		minorTicklen := majorTicklen / 2
		numMinorTicks := (r.xmax - r.xmin + 1) / r.xticksminor
		cr.SetLineWidth(1.0)
		tickSpacing := float64(width*r.xticksminor) / float64(r.xmax-r.xmin)
		halfBar := float64(width) / float64(r.xmax-r.xmin) / 2.0
		for i := 0; i <= numMinorTicks; i++ {
			xpos := float64(i)*tickSpacing + halfBar
			ypos := height - tickHeight
			cr.NewPath()
			cr.MoveTo(xpos, float64(ypos))
			cr.LineTo(xpos, float64(ypos)+float64(minorTicklen))
			cr.Stroke()
		}
		if tickHeight == 0 {
			tickHeight = minorTicklen
		}
	}
	if tickHeight > 0 {
		// draw a horizontal line
		ypos := height - tickHeight
		cr.NewPath()
		cr.MoveTo(0.0, float64(ypos))
		cr.LineTo(float64(width), float64(ypos))
		cr.Stroke()
	}

	return tickHeight
}

func (r *BarGraphRenderer) defaultDataProvider(_ int) ([]int, float64) {
	return r.data, 1.0
}

func (r *BarGraphRenderer) drawBars(cr *cairo.Context, width, height int) {
	if r.xmin == r.xmax {
		return
	}
	data, _ := r.barDataProvider(width)
	if len(data) == 0 {
		return
	}
	barWidth := float64(width+1) / float64(len(data))
	for i, val := range data {
		color := r.barColorFunc(i, val)
		left := float64(i) * barWidth
		if left > float64(width) {
			log.Printf("left edge %0.2f outside width %d", left, width)
		}
		barHeight := val * height / r.ymax
		top := height - barHeight
		cr.NewPath()
		cr.SetSourceRGB(color.red, color.green, color.blue)
		cr.Rectangle(left, float64(top), barWidth, float64(barHeight))
		cr.Fill()
	}
}

func (r *BarGraphRenderer) defaultColorFunc(_, _ int) Color {
	return r.fgcolor
}

type AveragingBarGraphRenderer struct {
	BarGraphRenderer
	minBarWidth int
}

func NewAveragingBarGraphRenderer(data []int) *AveragingBarGraphRenderer {
	rv := &AveragingBarGraphRenderer{
		BarGraphRenderer: BarGraphRenderer{
			data: data,
		},
		minBarWidth: 4,
	}
	rv.barDataProvider = rv.dataProvider
	return rv
}

func intMean(data []int) int {
	sum := 0
	if len(data) == 0 {
		return 0
	}
	for _, v := range data {
		sum += v
	}
	return sum / len(data)
}

func (r *AveragingBarGraphRenderer) dataProvider(width int) ([]int, float64) {
	dataLen := len(r.data)
	if dataLen == 0 {
		return r.data, 1
	}
	if width/dataLen >= r.minBarWidth {
		return r.data, 1
	}
	// this should round up rather than down
	// we should also find the smallest bar that works
	numGroups := (width + r.minBarWidth - 1) / r.minBarWidth
	groupSize := float64(dataLen) / float64(numGroups)
	rv := make([]int, numGroups)
	for i := 0; i < numGroups; i++ {
		groupStart := int(float64(i) * groupSize)
		groupEnd := int((float64(i) + 1.0) * groupSize)
		if groupEnd > dataLen {
			groupEnd = dataLen
		}
		group := r.data[groupStart:groupEnd]
		rv[i] = intMean(group)
	}
	return rv, groupSize
}

func MakeGradientFunc(low, high Color, max int) func(int, int) Color {
	maxf := float64(max)
	return func(_, val int) Color {
		point := float64(val) / maxf
		return Color{
			red:   low.red + (high.red-low.red)*point,
			green: low.green + (high.green-low.green)*point,
			blue:  low.blue + (high.blue-low.blue)*point,
		}
	}
}

var (
	_ GraphRenderer = &BarGraphRenderer{}
	_ GraphRenderer = &AveragingBarGraphRenderer{}
)
