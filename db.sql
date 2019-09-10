CREATE TABLE `rule` (
  `id` varchar(36) NOT NULL COMMENT 'rule规则ID',
  `path` varchar(128) NOT NULL COMMENT 'Mock API监听路径，支持正则表达式',
  `method` varchar(16) NOT NULL COMMENT 'Mock API请求方式，GET/POST/PATCH/PUT/DELETE等',
  `variable` blob COMMENT '规则级别的变量',
  `weight` blob COMMENT '规则级别的权重字段',
  `responses` blob COMMENT '规则对应的response regulation',
  `version` int(8) NOT NULL DEFAULT '0' COMMENT '规则版本号，每更新一次+1',
  `ctime` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '规则创建时间',
  `mtime` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '规则修改时间',
  `disabled` tinyint(1) NOT NULL DEFAULT '0' COMMENT '规则是否启用',
  PRIMARY KEY (`id`),
  UNIQUE KEY `rule_id_uindex` (`id`),
  UNIQUE KEY `rule_api_uindex` (`path`,`method`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4