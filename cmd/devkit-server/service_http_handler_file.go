package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/terrywh/devkit/app"
	"github.com/terrywh/devkit/entity"
	"github.com/terrywh/devkit/infra/log"
	"github.com/terrywh/devkit/stream"
)

type HttpFileHandler struct {
	app.HttpHandlerBase
}

func initHttpFileHandler(mux *http.ServeMux) *HttpFileHandler {
	handler := &HttpFileHandler{}
	mux.HandleFunc("/file/pull", handler.HandlePull)
	mux.HandleFunc("/file/push", handler.HandlePush)
	return handler
}

func (handler *HttpFileHandler) HandlePull(w http.ResponseWriter, r *http.Request) {
	bash_pid, _ := strconv.ParseUint(r.URL.Query().Get("bash_pid"), 10, 32)
	shell := DefaultShellHandler.find(int(bash_pid))
	if shell == nil {
		handler.Respond(w, entity.ErrSessionNotFound)
		return
	}
	src, err := stream.NewSessionStream(&shell.Server, shell.conn)
	if err != nil {
		handler.Respond(w, fmt.Errorf("stream file (conn): %w", err))
		return
	}
	defer src.CloseWrite()

	sf := entity.StreamFile{}
	io.WriteString(src, "/file/pull:")
	if err = app.SendJSON(src, sf); err != nil {
		handler.Respond(w, fmt.Errorf("stream file (req): %w", err))
		return
	}
	rsp := entity.Response{Error: &entity.DefaultErrorCode{}, Data: &sf}
	if err = app.ReadJSON(src.Reader(), &rsp); err != nil {
		handler.Respond(w, fmt.Errorf("stream file (rsp): %w", err))
		return
	}
	handler.Respond(w, sf)
	if sf.Target.Path != "" { // 指定了目标文件，直接写入文件
		proc := &app.StreamFile{Desc: &sf}
		err = proc.Do(context.Background(), src)
	} else { // 为指定时，在 RESPONSE 流中带回
		sf.Target.Size, err = io.Copy(w, src)
		if err == nil && sf.Target.Size != sf.Source.Size { // 将文件数据透传给 devctl 转写文件
			err = entity.ErrFileCorrupted
		}
	}
	if err != nil {
		log.Warn("<devkit-server> failed to stream file:", err)
	}
}

func (handler *HttpFileHandler) HandlePush(w http.ResponseWriter, r *http.Request) {
	bash_pid, _ := strconv.ParseUint(r.URL.Query().Get("bash_pid"), 10, 32)
	shell := DefaultShellHandler.find(int(bash_pid))
	if shell == nil {
		handler.Respond(w, entity.ErrSessionNotFound)
		return
	}
	src, err := stream.NewSessionStream(&shell.Server, shell.conn)
	if err != nil {
		handler.Respond(w, fmt.Errorf("stream file (conn): %w", err))
		return
	}
	defer src.CloseWrite()

	size, _ := strconv.ParseInt(r.URL.Query().Get("size"), 10, 64)
	perm, _ := strconv.ParseUint(r.URL.Query().Get("perm"), 8, 32)
	sf := entity.StreamFile{
		Source: entity.File{
			Path: r.URL.Query().Get("path"),
			Size: int64(size),
			Perm: uint32(perm),
		},
	}
	log.Info("<devkit-server> stream file:", sf.Source.Path)
	io.WriteString(src, "/file/push:")
	if err = app.SendJSON(src, sf); err != nil {
		handler.Respond(w, fmt.Errorf("stream file (req): %w", err))
		return
	}
	var file *os.File
	if file, err = os.Open(sf.Source.Path); err != nil {
		handler.Respond(w, fmt.Errorf("stream file (file): %w", err))
		return
	}
	if size, err = io.Copy(src, file); err != nil || size != sf.Source.Size {
		handler.Respond(w, fmt.Errorf("stream file (copy): %w", entity.ErrFileCorrupted))
		return
	}
	src.CloseWrite() // 文件发送完毕
	if err = app.Read(src.Reader(), &sf); err != nil {
		handler.Respond(w, fmt.Errorf("stream file (rsp): %w", err))
		return
	}
	handler.Respond(w, nil)
}
