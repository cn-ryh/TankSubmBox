package rest

import (
	"strings"

	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/result"
)

type TrackService struct {
	BaseBean
	trackDao *TrackDao
}

func (this *TrackService) Init() {
	this.BaseBean.Init()

	b := core.CONTEXT.GetBean(this.trackDao)
	if b, ok := b.(*TrackDao); ok {
		this.trackDao = b
	}
}

func (this *TrackService) CreateTrack(name, targetUserType, description string) (*Track, *result.WebResult) {
	if strings.TrimSpace(name) == "" {
		return nil, result.BadRequest("赛道名称不能为空")
	}

	if targetUserType != "STUDENT" && targetUserType != "TEACHER" && targetUserType != "BOTH" {
		return nil, result.BadRequest("目标用户类型必须是 STUDENT, TEACHER 或 BOTH")
	}

	existing := this.trackDao.FindByName(name)
	if existing != nil {
		return nil, result.BadRequest("赛道名称已存在")
	}

	track := &Track{
		Name:           name,
		TargetUserType: targetUserType,
		Description:    description,
	}

	track = this.trackDao.Create(track)
	return track, nil
}

func (this *TrackService) BulkCreateTracks(tracks []map[string]string) ([]Track, *result.WebResult) {
	var createdTracks []Track

	for _, trackData := range tracks {
		name := strings.TrimSpace(trackData["name"])
		if name == "" {
			continue
		}

		targetUserType := trackData["targetUserType"]
		if targetUserType != "STUDENT" && targetUserType != "TEACHER" && targetUserType != "BOTH" {
			targetUserType = "BOTH"
		}

		description := trackData["description"]

		existing := this.trackDao.FindByName(name)
		if existing != nil {
			continue
		}

		track := &Track{
			Name:           name,
			TargetUserType: targetUserType,
			Description:    description,
		}
		track = this.trackDao.Create(track)
		createdTracks = append(createdTracks, *track)
	}

	return createdTracks, nil
}

func (this *TrackService) GetAllTracks() []Track {
	return this.trackDao.FindAll()
}

func (this *TrackService) GetTracksByUserType(userType string) []Track {
	return this.trackDao.FindByUserType(userType)
}

func (this *TrackService) DeleteTrack(id int64) *result.WebResult {
	track := this.trackDao.Find(id)
	if track == nil {
		return result.BadRequest("赛道不存在")
	}

	this.trackDao.Delete(track)
	return nil
}