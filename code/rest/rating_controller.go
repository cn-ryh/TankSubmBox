package rest

import (
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/result"
	"net/http"
	"strconv"
)

type RatingController struct {
	BaseController
	ratingDao     *RatingDao
	submissionDao *SubmissionDao
}

func (this *RatingController) Init() {
	this.BaseController.Init()

	b := core.CONTEXT.GetBean(this.ratingDao)
	if b, ok := b.(*RatingDao); ok {
		this.ratingDao = b
	}

	b = core.CONTEXT.GetBean(this.submissionDao)
	if b, ok := b.(*SubmissionDao); ok {
		this.submissionDao = b
	}
}

func (this *RatingController) RegisterRoutes() map[string]func(writer http.ResponseWriter, request *http.Request) {
	routeMap := make(map[string]func(writer http.ResponseWriter, request *http.Request))
	routeMap["/api/rating/submit"] = this.Wrap(this.SubmitRating, USER_ROLE_JUDGE)
	routeMap["/api/rating/scored"] = this.Wrap(this.GetScoredSubmissions, USER_ROLE_JUDGE)
	return routeMap
}

// 提交评分
func (this *RatingController) SubmitRating(writer http.ResponseWriter, request *http.Request) *result.WebResult {
	user := this.checkUser(request)
	
	submissionIdStr := request.FormValue("submissionId")
	scoreStr := request.FormValue("score")
	comment := request.FormValue("comment")
	
	if submissionIdStr == "" || scoreStr == "" {
		return result.BadRequest("参数不完整")
	}
	
	submissionId, err := strconv.ParseInt(submissionIdStr, 10, 64)
	if err != nil {
		return result.BadRequest("submissionId格式错误")
	}
	
	score, err := strconv.Atoi(scoreStr)
	if err != nil || score < 0 || score > 100 {
		return result.BadRequest("评分必须在0-100之间")
	}
	
	// 检查提交是否存在
	submission := this.submissionDao.FindById(submissionId)
	if submission == nil {
		return result.BadRequest("提交作品不存在")
	}
	
	// 检查是否已经评分过
	existingRating := this.ratingDao.FindBySubmissionAndJudge(submissionId, user.Uuid)
	if existingRating != nil {
		// 更新现有评分
		existingRating.Score = score
		existingRating.Comment = comment
		this.ratingDao.Save(existingRating)
		return this.Success(existingRating)
	}
	
	// 创建新评分
	rating := &Rating{
		SubmissionId: submissionId,
		JudgeUuid:    user.Uuid,
		Score:        score,
		Comment:      comment,
	}
	
	this.ratingDao.Create(rating)
	return this.Success(rating)
}

// 获取已评分的提交作品
func (this *RatingController) GetScoredSubmissions(writer http.ResponseWriter, request *http.Request) *result.WebResult {
	user := this.checkUser(request)
	
	// 获取当前评委已经评分的所有提交ID
	scoredSubmissionIds := this.ratingDao.FindScoredSubmissionIdsByJudge(user.Uuid)
	
	return this.Success(scoredSubmissionIds)
}

