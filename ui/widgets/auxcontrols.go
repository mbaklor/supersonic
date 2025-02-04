package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/dweymouth/supersonic/backend"
	myTheme "github.com/dweymouth/supersonic/ui/theme"
	"github.com/dweymouth/supersonic/ui/util"
)

// The "aux" controls for playback, positioned to the right
// of the BottomPanel. Currently only volume control.
type AuxControls struct {
	widget.BaseWidget

	VolumeControl *VolumeControl
	loop          *miniButton

	container *fyne.Container
}

type miniButton struct {
	widget.Button
}

func newMiniButton(icon fyne.Resource) *miniButton {
	b := &miniButton{
		Button: widget.Button{
			Icon: icon,
		},
	}
	b.ExtendBaseWidget(b)
	return b
}

func (b *miniButton) MinSize() fyne.Size {
	return fyne.NewSize(24, 24)
}

func NewAuxControls(initialVolume int) *AuxControls {
	a := &AuxControls{
		VolumeControl: NewVolumeControl(initialVolume),
		loop:          newMiniButton(myTheme.RepeatIcon),
	}
	a.container = container.NewHBox(
		layout.NewSpacer(),
		container.NewVBox(
			util.NewHSpace(0), // hack to move everything down a tiny bit
			layout.NewSpacer(),
			a.VolumeControl,
			container.NewHBox(layout.NewSpacer(), a.loop, util.NewHSpace(5)),
			layout.NewSpacer(),
		),
	)
	return a
}

func (a *AuxControls) CreateRenderer() fyne.WidgetRenderer {
	a.ExtendBaseWidget(a)
	return widget.NewSimpleRenderer(a.container)
}

func (a *AuxControls) OnChangeLoopMode(f func()) {
	a.loop.OnTapped = f
}

func (a *AuxControls) SetLoopMode(mode backend.LoopMode) {
	switch mode {
	case backend.LoopModeAll:
		a.loop.Importance = widget.HighImportance
		a.loop.Icon = myTheme.RepeatIcon
	case backend.LoopModeOne:
		a.loop.Importance = widget.HighImportance
		a.loop.Icon = myTheme.RepeatOneIcon
	case backend.LoopModeNone:
		a.loop.Importance = widget.MediumImportance
		a.loop.Icon = myTheme.RepeatIcon
	}
	a.loop.Refresh()
}

type volumeSlider struct {
	widget.Slider

	Width float32
}

func NewVolumeSlider(width float32) *volumeSlider {
	v := &volumeSlider{
		Slider: widget.Slider{
			Min:         0,
			Max:         100,
			Step:        1,
			Orientation: widget.Horizontal,
			Value:       100,
		},
		Width: width,
	}
	v.ExtendBaseWidget(v)
	return v
}

func (v *volumeSlider) MinSize() fyne.Size {
	h := v.Slider.MinSize().Height
	return fyne.NewSize(v.Width, h)
}

func (v *volumeSlider) Scrolled(e *fyne.ScrollEvent) {
	v.SetValue(v.Value + float64(0.5*e.Scrolled.DY))
}

// This code will be OBSOLETE in Fyne 2.4
// which will natively add Tappable behavior to slider
// Tapped is called when a pointer tapped event is captured.
//
// Since: 2.4
func (v *volumeSlider) Tapped(e *fyne.PointEvent) {
	ratio := v.getRatio(e)
	val := v.Min + ratio*(v.Max-v.Min)
	v.SetValue(val)
	v.DragEnd()
}

func (v *volumeSlider) endOffset() float32 {
	return (theme.IconInlineSize()-4)/2 + theme.InnerPadding() - 1.5 // align with radio icons
}

func (v *volumeSlider) getRatio(e *fyne.PointEvent) float64 {
	pad := v.endOffset()

	x := e.Position.X
	y := e.Position.Y

	switch v.Orientation {
	case widget.Vertical:
		if y > v.Size().Height-pad {
			return 0.0
		} else if y < pad {
			return 1.0
		} else {
			return 1 - float64(y-pad)/float64(v.Size().Height-pad*2)
		}
	case widget.Horizontal:
		if x > v.Size().Width-pad {
			return 1.0
		} else if x < pad {
			return 0.0
		} else {
			return float64(x-pad) / float64(v.Size().Width-pad*2)
		}
	}
	return 0.0
}

type VolumeControl struct {
	widget.BaseWidget

	icon   *TappableIcon
	slider *volumeSlider

	OnSetVolume func(int)

	muted   bool
	lastVol int

	container *fyne.Container
}

func NewVolumeControl(initialVol int) *VolumeControl {
	v := &VolumeControl{}
	v.ExtendBaseWidget(v)
	v.icon = NewTappableIcon(theme.VolumeUpIcon())
	v.icon.OnTapped = v.toggleMute
	v.slider = NewVolumeSlider(100)
	v.lastVol = initialVol
	v.slider.Step = 1
	v.slider.Orientation = widget.Horizontal
	v.slider.Value = float64(v.lastVol)
	v.slider.OnChanged = v.onChanged
	v.container = container.NewHBox(container.NewCenter(v.icon), v.slider)
	return v
}

// Sets the volume that is displayed in the slider.
// Does not invoke OnSetVolume callback.
func (v *VolumeControl) SetVolume(vol int) {
	if (vol == v.lastVol && !v.muted) || (v.muted && vol == 0) {
		return
	}
	v.lastVol = vol
	v.muted = false
	v.setDisplayedVolume(vol)
}

func (v *VolumeControl) onChanged(volume float64) {
	vol := int(volume)
	v.lastVol = vol
	v.muted = false
	v.updateIconForVolume(vol)
	v.invokeOnVolumeChange(vol)
}

func (v *VolumeControl) toggleMute() {
	if !v.muted {
		v.muted = true
		v.lastVol = int(v.slider.Value)
		v.setDisplayedVolume(0)
		v.invokeOnVolumeChange(0)
	} else {
		v.muted = false
		v.setDisplayedVolume(v.lastVol)
		v.invokeOnVolumeChange(v.lastVol)
	}
}

func (v *VolumeControl) CreateRenderer() fyne.WidgetRenderer {
	v.ExtendBaseWidget(v)
	return widget.NewSimpleRenderer(v.container)
}

func (v *VolumeControl) setDisplayedVolume(vol int) {
	v.slider.Value = float64(vol)
	v.slider.Refresh()
	v.updateIconForVolume(vol)
}

func (v *VolumeControl) invokeOnVolumeChange(vol int) {
	if v.OnSetVolume != nil {
		v.OnSetVolume(vol)
	}
}

func (v *VolumeControl) updateIconForVolume(vol int) {
	if vol <= 0 {
		v.icon.Resource = theme.VolumeMuteIcon()
	} else if vol < 50 {
		v.icon.Resource = theme.VolumeDownIcon()
	} else {
		v.icon.Resource = theme.VolumeUpIcon()
	}
	v.icon.Refresh()
}
