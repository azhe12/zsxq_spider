package models

import "time"

type Like struct {
	ID        int    `gorm:"primary_key"`
	Ip        string `gorm:"type:varchar(20);not null;index:ip_idx"`
	Ua        string `gorm:"type:varchar(256);not null;"`
	Title     string `gorm:"type:varchar(128);not null;index:title_idx"`
	Hash      uint64 `gorm:"unique_index:hash_idx;"`
	CreatedAt time.Time
}
/*
CREATE TABLE zsxq.`topics` (
    `id` int(11) NOT NULL AUTO_INCREMENT,
    `group_id` int(11) NOT NULL DEFAULT '0' COMMENT '星球id',
    `topic_id` int(11) NOT NULL DEFAULT '0' COMMENT 'topic id',
    `topic_content` varchar(65535) NOT NULL DEFAULT '' COMMENT '原始topic内容',
    `topic_create_time` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00' COMMENT 'topic时间',
    `insert_time` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00' COMMENT '拉取时间',
    `modify_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_group_topic_id`(`group_id`, `topic_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='知识星球topic';
*/

type Topics struct {
	ID 				int 		`gorm:"primary_key"`
	GroupId 		int			`gorm:"type:int(11);not null;unique_index:idx_group_topic_id;"`
	TopicId 		int 		`gorm:"type:int(11);not null;"`
	TopicContent 	string 		`gorm:"type:varchar(65535);not null;"`
	TopicCreateTime time.Time 	`gorm:"type:timestamp"`
	CreateAt 		time.Time
	UpdateAt 		time.Time
}