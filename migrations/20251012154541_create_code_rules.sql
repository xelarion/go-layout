-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS code_rules
(
    id              BIGSERIAL PRIMARY KEY,
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at      BIGINT                   NOT NULL DEFAULT 0,
    created_by      BIGINT                   NOT NULL DEFAULT 0,
    updated_by      BIGINT                   NOT NULL DEFAULT 0,
    created_role_id BIGINT                   NOT NULL DEFAULT 0,
    target_form     VARCHAR(100)             NOT NULL,
    name            VARCHAR(100)             NOT NULL,
    slug            VARCHAR(100)             NOT NULL,
    prefix          VARCHAR(100)             NOT NULL,
    time_rule       smallint                 NOT NULL,
    generate_type   smallint                 NOT NULL,
    initial_serial  INT                      NOT NULL DEFAULT 1,
    serial_length   INT                      NOT NULL DEFAULT 6,
    step            INT                      NOT NULL DEFAULT 1
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_code_rules_slug ON code_rules (slug) WHERE deleted_at = 0;

COMMENT ON TABLE code_rules IS '编码规则';
COMMENT ON COLUMN code_rules.id IS '主键';
COMMENT ON COLUMN code_rules.created_at IS '创建时间';
COMMENT ON COLUMN code_rules.updated_at IS '更新时间';
COMMENT ON COLUMN code_rules.deleted_at IS '删除时间';
COMMENT ON COLUMN code_rules.created_by IS '创建人ID';
COMMENT ON COLUMN code_rules.updated_by IS '更新人ID';
COMMENT ON COLUMN code_rules.created_role_id IS '创建时角色ID';
COMMENT ON COLUMN code_rules.target_form IS '目标表单';
COMMENT ON COLUMN code_rules.name IS '编码名称';
COMMENT ON COLUMN code_rules.slug IS '编码标识(用于代码匹配)';
COMMENT ON COLUMN code_rules.prefix IS '编号前缀';
COMMENT ON COLUMN code_rules.time_rule IS '时间规则';
COMMENT ON COLUMN code_rules.generate_type IS '生成类型';
COMMENT ON COLUMN code_rules.initial_serial IS '初始流水号';
COMMENT ON COLUMN code_rules.serial_length IS '流水号长度';
COMMENT ON COLUMN code_rules.step IS '步长';

INSERT INTO code_rules (target_form, name, slug, prefix, time_rule, generate_type, initial_serial, serial_length, step)
VALUES ('组织管理', '组织编号', 'organization', 'ZZ', 0, 2, 1, 3, 1),
       ('岗位管理', '岗位编号', 'position', 'GW', 0, 2, 1, 4, 1),
       ('角色管理', '角色编号', 'role', 'JS', 0, 2, 1, 3, 1),
       ('生产计划', '计划编号', 'production_plan', 'JH', 3, 1, 1, 3, 1),
       ('生产工单', '工单编号', 'production_work_order', 'GD', 3, 1, 1, 4, 1),
       ('任务工单', '任务编号', 'production_work_order_task', 'RW', 3, 1, 1, 4, 1),
       ('工艺路线', '工艺路线编号', 'process_route', 'GYLX', 0, 2, 1, 3, 1),
       ('工序管理', '工序编号', 'process', 'GX', 0, 2, 1, 3, 1),
       ('车间管理', '车间编号', 'workshop', 'CJ', 0, 2, 1, 2, 1),
       ('工位管理', '工位编号', 'workstation', 'CZT', 0, 2, 1, 3, 1),
       ('出货检验', '出货检验单号', 'shipment_inspection', 'CHJY', 3, 1, 1, 3, 1),
       ('来料检验', '来料检验单号', 'incoming_inspection', 'LLJY', 3, 1, 1, 3, 1),
       ('过程检验', '过程检验单号', 'process_inspection', 'GCJY', 3, 1, 1, 3, 1),
       ('检验方案', '质检方案编号', 'inspection_plan', 'ZJFA', 0, 2, 1, 2, 1),
       ('检测项目', '检测项目编码', 'inspection_item', 'JCXM', 0, 2, 1, 2, 1),
       ('检测工具', '检测工具编码', 'inspection_tool', 'JCGJ', 0, 2, 1, 2, 1),
       ('常见缺陷', '缺陷编码', 'common_defect', 'QX', 0, 2, 1, 2, 1),
       ('缺陷类型', '缺陷类型编码', 'defect_type', 'QXLX', 0, 2, 1, 2, 1),
       ('缺陷位置', '缺陷位置编码', 'defect_location', 'QXWZ', 0, 2, 1, 2, 1),
       ('缺陷等级', '缺陷等级编码', 'defect_level', 'QXDJ', 0, 2, 1, 2, 1),
       ('仓库管理', '仓库编码', 'warehouse', 'CK', 0, 2, 1, 2, 1),
       ('库区管理', '库区编码', 'warehouse_area', 'KQ', 0, 2, 1, 2, 1),
       ('库位管理', '库位编码', 'warehouse_location', 'KW', 0, 2, 1, 2, 1),
       ('销售出库', '出库单编号', 'sales_outbound', 'CKD', 3, 1, 1, 3, 1),
       ('生产领料', '领料单编号', 'production_material_pick', 'LLD', 3, 1, 1, 3, 1),
       ('供应商退货', '退货单编号', 'supplier_return', 'THD', 3, 1, 1, 3, 1),
       ('产品入库', '入库单编号', 'product_inbound', 'RKD', 3, 1, 1, 3, 1),
       ('采购入库', '采购单编号', 'purchase_inbound', 'CGD', 3, 1, 1, 3, 1),
       ('生产退料', '退料单编号', 'production_return', 'TLD', 3, 1, 1, 3, 1),
       ('物料管理', '物料编号', 'material', 'WL', 0, 2, 1, 3, 1),
       ('产品管理', '产品编号', 'product', 'CP', 0, 2, 1, 3, 1),
       ('BOM管理', 'BOM编号', 'bom', 'BOM', 0, 2, 1, 3, 1),
       ('设备类型', '设备类型编号', 'equipment_type', 'SBLX', 0, 2, 1, 2, 1),
       ('设备台账', '设备编号', 'equipment', 'SB', 0, 2, 1, 2, 1),
       ('客户管理', '客户编号', 'customer', 'KH', 0, 2, 1, 3, 1),
       ('供应商管理', '供应商编号', 'supplier', 'GYS', 0, 2, 1, 3, 1);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS code_rules;
-- +goose StatementEnd
