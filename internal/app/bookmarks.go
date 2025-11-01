package app

import "encoding/json"

type Bookmarks struct {
	ws *WSManager
}

func NewBookmarks(ws *WSManager) *Bookmarks {
	return &Bookmarks{
		ws: ws,
	}
}

func (b *Bookmarks) AddBookmark(workspaceId string, line string, label string) error {
	w, err := b.ws.Get(workspaceId)
	if err != nil {
		return err
	}
	return w.AddBookmark(line, label)
}

func (b *Bookmarks) DeleteBookmark(workspaceId string, line string) error {
	w, err := b.ws.Get(workspaceId)
	if err != nil {
		return err
	}
	return w.DeleteBookmark(line)
}

func (b *Bookmarks) Bookmarks(workspaceId string) (string, error) {
	w, err := b.ws.Get(workspaceId)
	if err != nil {
		return "", err
	}
	bookmarks, err := w.Bookmarks()
	if err != nil {
		return "", err
	}
	bytes, err := json.Marshal(bookmarks)
	if err != nil {
		return "", err
	}
	s := string(bytes)
	return s, nil
}
