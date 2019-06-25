package models

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// QuestionnaireModel 问卷数据库
type QuestionnaireModel struct {
	Collection *mongo.Collection
}

// ProblemType 问题类型
type ProblemType string

// FillType 填空类型
type FillType string

// ProblemType 问题类型
const (
	ProblemNone   ProblemType = "none"   // 纯文字描述
	ProblemChoose ProblemType = "choose" // 选择题
	ProblemMatrix ProblemType = "matrix" // 矩阵题
	ProblemFill   ProblemType = "fill"   // 填空题
	ProblemScore  ProblemType = "score"  // 评分题
	ProblemSort   ProblemType = "sort"   // 排序题
)

// Fill 填空类型
const (
	FillPhone  FillType = "phone"  // 手机号码
	FillNumber FillType = "number" // 数字
	FillDate   FillType = "date"   // 日期
	FillEmail  FillType = "email"  // 电子邮箱
	FillID     FillType = "id"     // 身份证
	FillWeb    FillType = "web"    // 网址
	FillFile   FillType = "file"   // 文件(用于资料征集)
	FillAll    FillType = "all"    // 不限类型
)

// ProblemSchema 问卷问题
type ProblemSchema struct {
	Index   int64                // 题号
	Content string               // 问题本体
	Note    string               // 问题备注
	Type    ProblemType          // 问题类型

	// 选择题数据
	ChooseProblem struct {
		Options []struct {
			Index   int64              // 选项序号
			Content string             // 选项内容
		} // 问题选项
		MaxChoose int64 `bson:"max_choose" json:"max_choose"` // 最大可选项
	} `bson:"choose_problem,omitempty" json:"choose_problem,omitempty"`

	// 填空题数据
	FillProblem struct {
		Type      FillType `bson:"type"`                  // 填空类型
		MultiLine bool     `bson:"multi_line, omitempty" json:"multi_line"` // 是否多行
		MaxWord   int64    `bson:"max_word, omitempty" json:"max_word"`   // 最大字数
	} `bson:"fill_problem, omitempty" json:"fill_problem,omitempty"`

	// 评分题
	ScoreProblem struct {
		MinText string `bson:"min_text" json:"min_text"` // 低分描述（如：不满意，不重要，不愿意)
		MaxText string `bson:"max_text" json:"max_text"` // 高分描述
		Score   int64  `bson:"score"`    // 评分级别 1-x (最高为10)
	} `bson:"score_problem, omitempty" json:"score_problem,omitempty"`

	// 矩阵题
	MatrixProblem struct {
		Content []string // 题目
		Options []string // 选项
	} `bson:"matrix_problem, omitempty" json:"matrix_problem,omitempty"`

	// 排序题
	SortProblem struct {
		SortItem []struct {
			Index   int64  // 选项序号
			Content string // 选项内容
		} `bson:"sort_item" json:"sort_item"`
	} `bson:"sort_problem, omitempty" json:"sort_problem,omitempty"`
}

// StatisticsSchema 用户填写数据
type StatisticsSchema struct {
	UserID   primitive.ObjectID  `bson:"user_id" json:"user_id"` // 填写用户ID
	Data     []ProblemDataSchema // 问题答案
	Time     int64               // 提交时间
	Duration int64               // 花费时间(秒)
	IP       string         `bson:"ip" json:"ip"`     // 提交 IP 地址
}

// ProblemDataSchema 问题数据
type ProblemDataSchema struct {
	ProblemIndex int                `bson:"problem_index" json:"problem_index"`           // 题目序号
	StringValue  string             `bson:"string_value,omitempty" json:"string_value,omitempty"` // 填空题
	ChooseValue  []int              `bson:"choose_value,omitempty" json:"choose_value,omitempty"` // 选择题、矩阵题、排序题数据
	ScoreValue   int                `bson:"score_value,omitempty" json:"score_value,omitempty"`  // 评分题数据
	FileValue    primitive.ObjectID `bson:"file_value,omitempty" json:"file_value,omitempty"`   // 文件题数据
}

// QuestionnaireSchema 问卷数据结构
type QuestionnaireSchema struct {
	TaskID primitive.ObjectID `bson:"_id"` // 问卷所属任务ID [索引]

	Title       string             `bson:"title"`       // 问卷标题
	Description string             `bson:"description"` // 问卷描述
	Owner       primitive.ObjectID `bson:"owner"`       // 问卷创建者(冗余)
	Anonymous   bool               `bson:"anonymous"`   // 匿名收集
	Problems    []ProblemSchema    `bson:"problems"`    // 问卷问题
	Data        []StatisticsSchema `bson:"data"`        // 问题统计数据
}

// AddQuestionnaire 添加问卷
func (model *QuestionnaireModel) AddQuestionnaire(info QuestionnaireSchema) (primitive.ObjectID, error) {
	ctx, over := GetCtx()
	defer over()
	info.Problems = []ProblemSchema{}
	info.Data = []StatisticsSchema{}
	res, err := model.Collection.InsertOne(ctx, &info)
	if err != nil {
		return primitive.ObjectID{}, err
	}
	return res.InsertedID.(primitive.ObjectID), nil
}

// GetQuestionnaireInfoByID 获取问卷信息
func (model *QuestionnaireModel) GetQuestionnaireInfoByID(id primitive.ObjectID) (questionnaire QuestionnaireSchema, err error) {
	ctx, over := GetCtx()
	defer over()
	err = model.Collection.FindOne(ctx, bson.M{"_id": id}).Decode(&questionnaire)
	return
}

// SetQuestionnaireInfoByID 设置问卷信息
func (model *QuestionnaireModel) SetQuestionnaireInfoByID(id primitive.ObjectID, info QuestionnaireSchema) (err error) {
	ctx, over := GetCtx()
	defer over()
	updateItem := bson.M{}
	updateItem["title"] = info.Title
	updateItem["description"] = info.Description
	updateItem["anonymous"] = info.Anonymous
	res, err := model.Collection.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": updateItem}, options.Update().SetUpsert(true))
	if err != nil {
		return
	} else if res.MatchedCount < 1 {
		return ErrNotExist
	}
	return nil
}

// GetQuestionnaireQuestionsByID 获取问卷问题
func (model *QuestionnaireModel) GetQuestionnaireQuestionsByID(id primitive.ObjectID) (problems []ProblemSchema, err error) {
	ctx, over := GetCtx()
	defer over()
	var questionnaire QuestionnaireSchema
	err = model.Collection.FindOne(ctx, bson.M{"_id": id}).Decode(&questionnaire)
	if err != nil {
		return
	}
	problems = questionnaire.Problems
	return
}

// SetQuestionnaireQuestionsByID 修改问卷问题
func (model *QuestionnaireModel) SetQuestionnaireQuestionsByID(id primitive.ObjectID, questions []ProblemSchema) (err error) {
	ctx, over := GetCtx()
	defer over()
	updateItem := bson.M{
		"problems": questions,
	}
	res, err := model.Collection.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": updateItem})
	if err != nil {
		return
	} else if res.MatchedCount < 1 {
		return ErrNotExist
	}
	return nil
}

// GetQuestionnaireAnswersByID 获取问卷答案数据
func (model *QuestionnaireModel) GetQuestionnaireAnswersByID(id primitive.ObjectID) (answers []StatisticsSchema, err error) {
	ctx, over := GetCtx()
	defer over()
	var questionnaire QuestionnaireSchema
	err = model.Collection.FindOne(ctx, bson.M{"_id": id}).Decode(&questionnaire)
	if err != nil {
		return
	}
	answers = questionnaire.Data
	return
}

// AddAnswer 添加新回答
func (model *QuestionnaireModel) AddAnswer(id primitive.ObjectID, statistics StatisticsSchema) (err error) {
	ctx, over := GetCtx()
	defer over()
	res, err := model.Collection.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$push": bson.M{"data": statistics} })
	if err != nil {
		return
	} else if res.MatchedCount < 1 {
		return ErrNotExist
	}
	return nil
}

// GetAnswerByUserID 根据用户获取答案
func (model *QuestionnaireModel) GetAnswerByUserID(id, userID primitive.ObjectID) (StatisticsSchema, error) {
	ctx, over := GetCtx()
	defer over()
	res := QuestionnaireSchema{}
	err := model.Collection.FindOne(ctx,
		bson.M{"_id": id, "data.user_id": userID}, options.FindOne().SetProjection(bson.M{"data": 1})).Decode(&res)
	if err != nil {
		return StatisticsSchema{}, err
	} else if len(res.Data) < 1 {
		return StatisticsSchema{}, ErrNotExist
	}
	return res.Data[0], nil
}