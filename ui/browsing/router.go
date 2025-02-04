package browsing

import (
	"github.com/dweymouth/supersonic/backend"
	"github.com/dweymouth/supersonic/ui/controller"
	"github.com/dweymouth/supersonic/ui/util"
)

type NavigationHandler interface {
	SetPage(Page)
	CurrentPage() controller.Route
}

type Router struct {
	App        *backend.App
	Controller *controller.Controller
	Nav        NavigationHandler
	widgetPool *util.WidgetPool
}

func NewRouter(app *backend.App, controller *controller.Controller, nav NavigationHandler) Router {
	r := Router{
		App:        app,
		Controller: controller,
		Nav:        nav,
		widgetPool: util.NewWidgetPool(),
	}
	return r
}

func (r Router) CreatePage(rte controller.Route) Page {
	switch rte.Page {
	case controller.Album:
		return NewAlbumPage(rte.Arg, &r.App.Config.AlbumPage, r.widgetPool, r.App.PlaybackManager, r.App.ServerManager.Server, r.App.ImageManager, r.Controller)
	case controller.Albums:
		return NewAlbumsPage(&r.App.Config.AlbumsPage, r.widgetPool, r.Controller, r.App.PlaybackManager, r.App.ServerManager.Server, r.App.ImageManager)
	case controller.Artist:
		return NewArtistPage(rte.Arg, &r.App.Config.ArtistPage, r.widgetPool, r.App.PlaybackManager, r.App.ServerManager.Server, r.App.ImageManager, r.Controller)
	case controller.Artists:
		return NewArtistsPage(r.Controller, r.widgetPool, r.App.PlaybackManager, r.App.ServerManager.Server, r.App.ImageManager)
	case controller.Favorites:
		return NewFavoritesPage(&r.App.Config.FavoritesPage, r.widgetPool, r.Controller, r.App.ServerManager.Server, r.App.PlaybackManager, r.App.ImageManager)
	case controller.Genre:
		return NewGenrePage(rte.Arg, r.widgetPool, r.Controller, r.App.PlaybackManager, r.App.ServerManager.Server, r.App.ImageManager)
	case controller.Genres:
		return NewGenresPage(r.Controller, r.App.ServerManager.Server)
	case controller.NowPlaying:
		return NewNowPlayingPage(rte.Arg, r.Controller, r.widgetPool, &r.App.Config.NowPlayingPage, r.App.PlaybackManager, r.App.Player)
	case controller.Playlist:
		return NewPlaylistPage(rte.Arg, &r.App.Config.PlaylistPage, r.widgetPool, r.Controller, r.App.ServerManager, r.App.PlaybackManager, r.App.ImageManager)
	case controller.Playlists:
		return NewPlaylistsPage(r.Controller, r.widgetPool, &r.App.Config.PlaylistsPage, r.App.ServerManager.Server)
	case controller.Tracks:
		return NewTracksPage(r.Controller, &r.App.Config.TracksPage, r.widgetPool, r.App.ServerManager.Server)
	}
	return nil
}

func (r Router) NavigateTo(rte controller.Route) {
	if rte != r.Nav.CurrentPage() {
		r.Nav.SetPage(r.CreatePage(rte))
	}
}
