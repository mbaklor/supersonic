package browsing

import (
	"log"
	"strconv"
	"strings"

	"github.com/dweymouth/supersonic/backend/mediaprovider"
	"github.com/dweymouth/supersonic/sharedutil"
	"github.com/dweymouth/supersonic/ui/controller"
	"github.com/dweymouth/supersonic/ui/layouts"
	"github.com/dweymouth/supersonic/ui/widgets"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var _ fyne.Widget = (*ArtistPage)(nil)

type GenresPage struct {
	widget.BaseWidget

	contr  *controller.Controller
	mp     mediaprovider.MediaProvider
	genres []*mediaprovider.Genre
	list   *GenreList

	titleDisp *widget.RichText
	container *fyne.Container
	searcher  *widgets.SearchEntry
}

func NewGenresPage(contr *controller.Controller, mp mediaprovider.MediaProvider) *GenresPage {
	return newGenresPage(contr, mp, "")
}

func newGenresPage(contr *controller.Controller, mp mediaprovider.MediaProvider, searchText string) *GenresPage {
	a := &GenresPage{
		contr:     contr,
		mp:        mp,
		titleDisp: widget.NewRichTextWithText("Genres"),
	}
	a.ExtendBaseWidget(a)
	a.titleDisp.Segments[0].(*widget.TextSegment).Style.SizeName = theme.SizeNameHeadingText
	a.list = NewGenreList(nil)
	a.list.OnNavTo = func(id string) { a.contr.NavigateTo(controller.GenreRoute(id)) }
	a.searcher = widgets.NewSearchEntry()
	a.searcher.OnSearched = a.onSearched
	a.searcher.Entry.Text = searchText
	a.buildContainer()
	go a.load(searchText != "")
	return a
}

// should be called asynchronously
func (a *GenresPage) load(searchOnLoad bool) {
	genres, err := a.mp.GetGenres()
	if err != nil {
		log.Printf("error loading genres: %v", err.Error())
	}
	a.genres = genres
	if searchOnLoad {
		a.onSearched(a.searcher.Entry.Text)
	} else {
		a.list.Items = genres
		a.list.Refresh()
	}
}

func (a *GenresPage) onSearched(query string) {
	// since the artists and genres lists are returned in full non-paginated, we will do our own
	// simple search based on the artist/genre name, rather than calling a server API
	if query == "" {
		a.list.Items = a.genres
	} else {
		query = strings.ToLower(query)
		result := sharedutil.FilterSlice(a.genres, func(x *mediaprovider.Genre) bool {
			return strings.Contains(strings.ToLower(x.Name), query)
		})
		a.list.Items = result
	}
	a.list.Refresh()
}

var _ Searchable = (*GenresPage)(nil)

func (a *GenresPage) SearchWidget() fyne.Focusable {
	return a.searcher
}

func (a *GenresPage) Route() controller.Route {
	return controller.GenresRoute()
}

func (a *GenresPage) Reload() {
	go a.load(false)
}

func (a *GenresPage) Save() SavedPage {
	return &savedArtistsGenresPage{
		contr:      a.contr,
		mp:         a.mp,
		searchText: a.searcher.Entry.Text,
	}
}

type savedArtistsGenresPage struct {
	isGenresPage bool
	contr        *controller.Controller
	mp           mediaprovider.MediaProvider
	searchText   string
}

func (s *savedArtistsGenresPage) Restore() Page {
	return newGenresPage(s.contr, s.mp, s.searchText)
}

func (a *GenresPage) buildContainer() {
	searchVbox := container.NewVBox(layout.NewSpacer(), a.searcher, layout.NewSpacer())
	a.container = container.New(&layouts.MaxPadLayout{PadLeft: 15, PadRight: 15, PadTop: 5, PadBottom: 15},
		container.NewBorder(
			container.New(&layouts.MaxPadLayout{PadLeft: -5},
				container.NewHBox(a.titleDisp, layout.NewSpacer(), searchVbox)),
			nil, nil, nil, a.list))
}

func (a *GenresPage) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(a.container)
}

type GenreList struct {
	widget.BaseWidget

	Items   []*mediaprovider.Genre
	OnNavTo func(string)

	columnsLayout *layouts.ColumnsLayout
	hdr           *widgets.ListHeader
	list          *widget.List
	container     *fyne.Container
}

type GenreListRow struct {
	widget.BaseWidget

	Item     *mediaprovider.Genre
	OnTapped func()

	nameLabel       *widget.Label
	albumCountLabel *widget.Label
	trackCountLabel *widget.Label

	container *fyne.Container
}

func NewGenreListRow(layout *layouts.ColumnsLayout) *GenreListRow {
	a := &GenreListRow{
		nameLabel:       widget.NewLabel(""),
		albumCountLabel: widget.NewLabel(""),
		trackCountLabel: widget.NewLabel(""),
	}
	a.ExtendBaseWidget(a)
	a.albumCountLabel.Alignment = fyne.TextAlignTrailing
	a.trackCountLabel.Alignment = fyne.TextAlignTrailing
	a.container = container.New(layout, a.nameLabel, a.albumCountLabel, a.trackCountLabel)
	return a
}

func NewGenreList(items []*mediaprovider.Genre) *GenreList {
	a := &GenreList{
		Items:         items,
		columnsLayout: layouts.NewColumnsLayout([]float32{-1, 125, 125}),
	}
	a.ExtendBaseWidget(a)
	a.hdr = widgets.NewListHeader([]widgets.ListColumn{
		{"Name", fyne.TextAlignLeading, false}, {"Album Count", fyne.TextAlignTrailing, false}, {"Track Count", fyne.TextAlignTrailing, false}}, a.columnsLayout)
	a.hdr.DisableSorting = true
	a.list = widget.NewList(
		func() int { return len(a.Items) },
		func() fyne.CanvasObject {
			r := NewGenreListRow(a.columnsLayout)
			r.OnTapped = func() { a.onRowDoubleTapped(r.Item) }
			return r
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			row := item.(*GenreListRow)
			row.Item = a.Items[id]
			row.nameLabel.Text = row.Item.Name
			row.albumCountLabel.Text = strconv.Itoa(row.Item.AlbumCount)
			row.trackCountLabel.Text = strconv.Itoa(row.Item.TrackCount)
			row.Refresh()
		},
	)
	a.container = container.NewBorder(a.hdr, nil, nil, nil, a.list)
	return a
}

func (a *GenreList) onRowDoubleTapped(item *mediaprovider.Genre) {
	if a.OnNavTo != nil {
		a.OnNavTo(item.Name)
	}
}

func (a *GenreListRow) Tapped(*fyne.PointEvent) {
	if a.OnTapped != nil {
		a.OnTapped()
	}
}

func (a *GenreList) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(a.container)
}

func (a *GenreListRow) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(a.container)
}
