package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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
	Images  []primitive.ObjectID // 图片附件ID

	// 选择题数据
	ChooseProblem struct {
		Options []struct {
			Index   int64              // 选项序号
			Content string             // 选项内容
			Image   primitive.ObjectID `bson:"image,omitempty"` // 图片附件ID
		} `bson:"options"` // 问题选项
		MaxChoose int64 `bson:"max_choose"` // 最大可选项
	} `bson:"choose_problem,omitempty"`

	// 填空题数据
	FillProblem struct {
		Type      FillType `bson:"type"`                  // 填空类型
		MultiLine bool     `bson:"multi_line, omitempty"` // 是否多行
		MaxWord   int64    `bson:"max_word, omitempty"`   // 最大字数
	} `bson:"fill_problem, omitempty"`

	// 评分题
	ScoreProblem struct {
		MinText string `bson:"min_text"` // 低分描述（如：不满意，不重要，不愿意)
		MaxText string `bson:"max_text"` // 高分描述
		Score   int64  `bson:"score"`    // 评分级别 1-x (最高为10)
	} `bson:"score_problem, omitempty"`

	// 矩阵题
	MatrixProblem struct {
		Content []string // 题目
		Options []string // 选项
	} `bson:"matrix_problem, omitempty"`

	// 排序题
	SortProblem struct {
		SortItem []struct {
			Index   int64  // 选项序号
			Content string // 选项内容
		} `bson:"sort_item"`
	} `bson:"sort_problem, omitempty"`
}

// StatisticsSchema 用户填写数据
type StatisticsSchema struct {
	UserID   primitive.ObjectID  `bson:"user_id"` // 填写用户ID
	Data     []ProblemDataSchema // 问题答案
	Time     int64               // 提交时间
	Duration int64               // 花费时间(秒)
	IP       string              // 提交 IP 地址
}

// ProblemDataSchema 问题数据
type ProblemDataSchema struct {
	ProblemIndex int                `bson:"problem_index"`           // 题目序号
	StringValue  string             `bson:"string_value, omitempty"` // 填空题
	ChooseValue  []int              `bson:"choose_value, omitempty"` // 选择题、矩阵题、排序题数据
	ScoreValue   int                `bson:"score_value, omitempty"`  // 评分题数据
	FileValue    primitive.ObjectID `bson:"file_value, omitempty"`   // 文件题数据
}

// QuestionnaireSchema 问卷数据结构
type QuestionnaireSchema struct {
	TaskID primitive.ObjectID `bson:"_id"` // 问卷所属任务ID [索引]

	Title       string             // 问卷标题
	Description string             // 问卷描述
	Owner       string             // 问卷创建者(冗余)
	Anonymous   bool               // 匿名收集
	Problems    []ProblemSchema    // 问卷问题
	Data        []StatisticsSchema // 问题统计数据
}
