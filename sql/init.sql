-- ============================================================
-- 客户升级记录管理系统 — 数据库初始化
-- ============================================================
CREATE DATABASE IF NOT EXISTS upgrade_tracker
  DEFAULT CHARACTER SET utf8mb4
  DEFAULT COLLATE utf8mb4_unicode_ci;

USE upgrade_tracker;

CREATE TABLE IF NOT EXISTS clients (
  id              INT AUTO_INCREMENT PRIMARY KEY,
  name            VARCHAR(100) NOT NULL        COMMENT '客户名称',
  type            VARCHAR(20)  NOT NULL DEFAULT 'other' COMMENT '行业: finance/retail/medical/edu/gov/other',
  contact         VARCHAR(50)                  COMMENT '联系人',
  note            TEXT                         COMMENT '备注',
  current_version VARCHAR(50)  NOT NULL DEFAULT 'v1.0.0' COMMENT '当前版本',
  created_at      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='客户表';

CREATE TABLE IF NOT EXISTS upgrade_records (
  id           INT AUTO_INCREMENT PRIMARY KEY,
  client_id    INT          NOT NULL            COMMENT '客户ID',
  version      VARCHAR(50)  NOT NULL            COMMENT '版本号',
  upgrade_date DATE         NOT NULL            COMMENT '升级日期',
  operator     VARCHAR(50)  NOT NULL DEFAULT '' COMMENT '操作人',
  tags         VARCHAR(200) NOT NULL DEFAULT '' COMMENT '变更类型，逗号分隔',
  description  TEXT                             COMMENT '升级说明',
  files        TEXT                             COMMENT '关联文件，逗号分隔',
  created_at   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_client (client_id),
  INDEX idx_date   (upgrade_date),
  FOREIGN KEY (client_id) REFERENCES clients(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='升级记录表';

CREATE TABLE IF NOT EXISTS product_images (
  id           INT AUTO_INCREMENT PRIMARY KEY,
  name         VARCHAR(100) NOT NULL        COMMENT '产品/镜像名称',
  version      VARCHAR(50)  NOT NULL        COMMENT '版本号',
  type         VARCHAR(50)  NOT NULL DEFAULT 'docker' COMMENT '类型，如 docker, tarpkg',
  public_url   TEXT                         COMMENT '公网地址',
  internal_url TEXT                         COMMENT '内网地址',
  config_guide TEXT                         COMMENT '配置指导说明',
  description  TEXT                         COMMENT '附加说明',
  created_at   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_name_version (name, version)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='产品镜像表';

INSERT INTO clients (name, type, contact, note, current_version, created_at) VALUES
  ('星汇银行',     'finance', '张磊',   '核心银行系统', 'v3.2.1', '2024-01-10 00:00:00'),
  ('美达零售集团', 'retail',  '李佳',   'ERP升级项目',  'v2.5.0', '2024-02-15 00:00:00'),
  ('仁爱医院',     'medical', '王医生', 'HIS系统',      'v1.8.3', '2024-03-01 00:00:00');

INSERT INTO upgrade_records (client_id, version, upgrade_date, operator, tags, description, files) VALUES
  (1,'v3.2.1','2025-03-10','陈运维','feature,fix','优化了转账模块性能，修复了报表导出格式异常的问题，新增了多币种结算支持功能。','core-bank-3.2.1.jar,config-update.xml'),
  (1,'v3.1.0','2025-01-20','陈运维','feature','升级至3.1.0版本，新增实时风控模块，优化了账户管理界面交互体验。','core-bank-3.1.0.jar'),
  (1,'v3.0.5','2024-11-08','李工','hotfix','紧急修复日结汇总时的精度丢失问题，影响范围：大额转账汇总报表。','patch-3.0.5.jar'),
  (2,'v2.5.0','2025-03-01','赵运维','feature,config','全面升级至2.5版本，引入新供应链管理模块，调整仓储策略配置参数。','erp-2.5.0.war,supply-chain-config.yaml'),
  (2,'v2.4.2','2025-01-12','赵运维','fix','修复了门店库存同步延迟问题，优化数据库连接池配置。','erp-patch-2.4.2.war'),
  (3,'v1.8.3','2025-02-18','孙工','fix,config','修复了挂号排队逻辑的并发问题，更新了医保接口对接配置。','his-1.8.3.jar,yibao-config.properties');
