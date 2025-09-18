package rest

import (
	"strings"

	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/result"
)

type CollegeService struct {
	BaseBean
	collegeDao *CollegeDao
}

func (this *CollegeService) Init() {
	this.BaseBean.Init()

	b := core.CONTEXT.GetBean(this.collegeDao)
	if b, ok := b.(*CollegeDao); ok {
		this.collegeDao = b
	}
}

func (this *CollegeService) CreateCollege(name string) (*College, *result.WebResult) {
	if strings.TrimSpace(name) == "" {
		return nil, result.BadRequest("学院名称不能为空")
	}

	existing := this.collegeDao.FindByName(name)
	if existing != nil {
		return nil, result.BadRequest("学院名称已存在")
	}

	college := &College{
		Name: name,
	}

	college = this.collegeDao.Create(college)
	return college, nil
}

func (this *CollegeService) BulkCreateColleges(names []string) ([]College, *result.WebResult) {
	var createdColleges []College

	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}

		existing := this.collegeDao.FindByName(name)
		if existing != nil {
			continue
		}

		college := &College{
			Name: name,
		}
		college = this.collegeDao.Create(college)
		createdColleges = append(createdColleges, *college)
	}

	return createdColleges, nil
}

func (this *CollegeService) GetAllColleges() []College {
	return this.collegeDao.FindAll()
}

func (this *CollegeService) DeleteCollege(id int64) *result.WebResult {
	college := this.collegeDao.Find(id)
	if college == nil {
		return result.BadRequest("学院不存在")
	}

	this.collegeDao.Delete(college)
	return nil
}