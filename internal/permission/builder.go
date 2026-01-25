package permission

import (
	"context"
)

// Builder provides methods to build the permission tree
type Builder struct{}

// NewBuilder creates a new permission tree builder
func NewBuilder() *Builder {
	return &Builder{}
}

// GetTree returns the permission tree
func (b *Builder) GetTree(ctx context.Context) ([]*Node, error) {
	// Define all permissions, grouped by module
	permissions := []*Node{
		{
			ID:   "menu:production_management",
			Name: "生产管理",
			Children: []*Node{
				{
					ID:   "menu:production_plan",
					Name: "生产计划管理",
					Children: []*Node{
						{ID: ProductionPlanList, Name: "列表"},
						{ID: ProductionPlanDetail, Name: "详情"},
						{ID: ProductionPlanCreate, Name: "新建"},
						{ID: ProductionPlanUpdate, Name: "编辑"},
						{ID: ProductionPlanDelete, Name: "删除"},
						{ID: ProductionPlanExport, Name: "导出"},
					},
				},
				{
					ID:   "menu:process_route",
					Name: "工艺路线管理",
					Children: []*Node{
						{ID: ProcessRouteList, Name: "列表"},
						{ID: ProcessRouteDetail, Name: "详情"},
						{ID: ProcessRouteCreate, Name: "新建"},
						{ID: ProcessRouteUpdate, Name: "编辑"},
						{ID: ProcessRouteDelete, Name: "删除"},
						{ID: ProcessRouteExport, Name: "导出"},
					},
				},
				{
					ID:   "menu:process",
					Name: "工序管理",
					Children: []*Node{
						{ID: ProcessList, Name: "列表"},
						{ID: ProcessCreate, Name: "新建"},
						{ID: ProcessUpdate, Name: "编辑"},
						{ID: ProcessDelete, Name: "删除"},
						{ID: ProcessExport, Name: "导出"},
					},
				},
			},
		},

		{
			ID:   "menu:product_material",
			Name: "产品物料",
			Children: []*Node{
				{
					ID:   "menu:material",
					Name: "物料管理",
					Children: []*Node{
						{ID: MaterialList, Name: "列表"},
						{ID: MaterialDetail, Name: "详情"},
						{ID: MaterialCreate, Name: "新建"},
						{ID: MaterialUpdate, Name: "编辑"},
						{ID: MaterialUpdateEnabled, Name: "更新状态"},
						{ID: MaterialDelete, Name: "删除"},
						{ID: MaterialExport, Name: "导出"},
					},
				},
				{
					ID:   "menu:product",
					Name: "产品管理",
					Children: []*Node{
						{ID: ProductList, Name: "列表"},
						{ID: ProductDetail, Name: "详情"},
						{ID: ProductCreate, Name: "新建"},
						{ID: ProductUpdate, Name: "编辑"},
						{ID: ProductUpdateEnabled, Name: "更新状态"},
						{ID: ProductDelete, Name: "删除"},
						{ID: ProductExport, Name: "导出"},
					},
				},
				{
					ID:   "menu:bom",
					Name: "BOM管理",
					Children: []*Node{
						{ID: BomList, Name: "列表"},
						{ID: BomDetail, Name: "详情"},
						{ID: BomCreate, Name: "新建"},
						{ID: BomUpdate, Name: "编辑"},
						{ID: BomDelete, Name: "删除"},
					},
				},
			},
		},
		{
			ID:   "menu:system_management",
			Name: "系统管理",
			Children: []*Node{
				{
					ID:   "menu:system_settings",
					Name: "系统设置",
					Children: []*Node{
						{
							ID:   "menu:code_rule",
							Name: "编码规则",
							Children: []*Node{
								{ID: CodeRuleList, Name: "列表"},
								{ID: CodeRuleUpdate, Name: "编辑"},
							},
						},
						{
							ID:   "menu:unit",
							Name: "单位管理",
							Children: []*Node{
								{ID: UnitList, Name: "列表"},
								{ID: UnitDetail, Name: "详情"},
								{ID: UnitCreate, Name: "新建"},
								{ID: UnitUpdate, Name: "编辑"},
								{ID: UnitDelete, Name: "删除"},
							},
						},
						{
							ID:   "menu:announcement",
							Name: "通知公告",
							Children: []*Node{
								{ID: AnnouncementList, Name: "列表"},
								{ID: AnnouncementDetail, Name: "查看"},
								{ID: AnnouncementCreate, Name: "新建"},
								{ID: AnnouncementUpdate, Name: "编辑"},
								{ID: AnnouncementDelete, Name: "删除"},
							},
						},
					},
				},
				{
					ID:   "menu:system_query",
					Name: "系统查询",
					Children: []*Node{
						{
							ID:   "menu:batch_record",
							Name: "批量操作记录",
							Children: []*Node{
								{ID: BatchRecordList, Name: "列表"},
								{ID: BatchRecordDetail, Name: "详情"},
							},
						},
					},
				},
				{
					ID:   "menu:organization_management",
					Name: "组织机构",
					Children: []*Node{
						{
							ID:   "menu:organization",
							Name: "组织管理",
							Children: []*Node{
								{ID: OrganizationList, Name: "列表"},
								{ID: OrganizationCreate, Name: "新建"},
								{ID: OrganizationUpdate, Name: "编辑"},
								{ID: OrganizationDelete, Name: "删除"},
							},
						},
						{
							ID:   "menu:position",
							Name: "岗位管理",
							Children: []*Node{
								{ID: PositionList, Name: "列表"},
								{ID: PositionCreate, Name: "新建"},
								{ID: PositionUpdate, Name: "编辑"},
								{ID: PositionDelete, Name: "删除"},
							},
						},
						{
							ID:   "menu:user",
							Name: "用户管理",
							Children: []*Node{
								{ID: UserList, Name: "列表"},
								{ID: UserCreate, Name: "新建"},
								{ID: UserUpdate, Name: "编辑"},
								{ID: UserUpdateEnabled, Name: "生效/失效"},
								{ID: UserUpdatePassword, Name: "修改密码"},
								{ID: UserDelete, Name: "删除"},
								{ID: UserExport, Name: "导出"},
							},
						},
						{
							ID:   "menu:role",
							Name: "角色管理",
							Children: []*Node{
								{ID: RoleList, Name: "列表"},
								{ID: RoleCreate, Name: "新建"},
								{ID: RoleUpdate, Name: "编辑"},
								{ID: RoleUpdateEnabled, Name: "生效/失效"},
								{ID: RoleDelete, Name: "删除"},
								{ID: PermissionUpdate, Name: "权限设置"},
							},
						},
					},
				},
			},
		},
		{
			ID:   "menu:archive_management",
			Name: "档案管理",
			Children: []*Node{
				{
					ID:   "menu:supplier",
					Name: "供应商管理",
					Children: []*Node{
						{ID: SupplierList, Name: "列表"},
						{ID: SupplierDetail, Name: "详情"},
						{ID: SupplierCreate, Name: "新建"},
						{ID: SupplierUpdate, Name: "编辑"},
						{ID: SupplierDelete, Name: "删除"},
						{ID: SupplierImport, Name: "导入"},
						{ID: SupplierExport, Name: "导出"},
					},
				},
				{
					ID:   "menu:customer",
					Name: "客户管理",
					Children: []*Node{
						{ID: CustomerList, Name: "列表"},
						{ID: CustomerDetail, Name: "详情"},
						{ID: CustomerCreate, Name: "新建"},
						{ID: CustomerUpdate, Name: "编辑"},
						{ID: CustomerDelete, Name: "删除"},
						{ID: CustomerImport, Name: "导入"},
						{ID: CustomerExport, Name: "导出"},
					},
				},
				{
					ID:   "menu:equipment",
					Name: "设备台账",
					Children: []*Node{
						{ID: EquipmentList, Name: "列表"},
						{ID: EquipmentDetail, Name: "详情"},
						{ID: EquipmentCreate, Name: "新建"},
						{ID: EquipmentUpdate, Name: "编辑"},
						{ID: EquipmentDelete, Name: "删除"},
						{ID: EquipmentExport, Name: "导出"},
						{ID: EquipmentImport, Name: "导入"},
					},
				},
				{
					ID:   "menu:equipment_type",
					Name: "设备类型管理",
					Children: []*Node{
						{ID: EquipmentTypeList, Name: "列表"},
						{ID: EquipmentTypeDetail, Name: "详情"},
						{ID: EquipmentTypeCreate, Name: "新建"},
						{ID: EquipmentTypeUpdate, Name: "编辑"},
						{ID: EquipmentTypeDelete, Name: "删除"},
						{ID: EquipmentTypeImport, Name: "导入"},
					},
				},
			},
		}, {
			ID:   "menu:warehouse_management",
			Name: "仓储管理",
			Children: []*Node{
				{
					ID:   "menu:warehouse",
					Name: "仓库管理",
					Children: []*Node{
						{
							ID:   "menu:warehouse",
							Name: "仓库管理",
							Children: []*Node{
								{ID: WarehouseCreate, Name: "新建"},
								{ID: WarehouseUpdate, Name: "编辑"},
								{ID: WarehouseDelete, Name: "删除"},
								{ID: WarehouseDetail, Name: "详情"},
								{ID: WarehouseList, Name: "列表"},
							},
						},
						{
							ID:   "menu:warehouse_area",
							Name: "库区管理",
							Children: []*Node{
								{ID: WarehouseAreaCreate, Name: "新建"},
								{ID: WarehouseAreaUpdate, Name: "编辑"},
								{ID: WarehouseAreaDelete, Name: "删除"},
								{ID: WarehouseAreaDetail, Name: "详情"},
								{ID: WarehouseAreaList, Name: "列表"},
							},
						},
						{
							ID:   "menu:warehouse_location",
							Name: "库位管理",
							Children: []*Node{
								{ID: WarehouseLocationCreate, Name: "新建"},
								{ID: WarehouseLocationUpdate, Name: "编辑"},
								{ID: WarehouseLocationDelete, Name: "删除"},
								{ID: WarehouseLocationDetail, Name: "详情"},
								{ID: WarehouseLocationList, Name: "列表"},
							},
						},
					},
				},
			},
		},
		{
			ID:   "menu:quality_management",
			Name: "质量管理",
			Children: []*Node{
				{
					ID:   "menu:inspection_management",
					Name: "检测管理",
					Children: []*Node{
						{
							ID:   "menu:inspection_tool",
							Name: "检测工具",
							Children: []*Node{
								{ID: InspectionToolCreate, Name: "新建"},
								{ID: InspectionToolUpdate, Name: "编辑"},
								{ID: InspectionToolDelete, Name: "删除"},
								{ID: InspectionToolDetail, Name: "详情"},
								{ID: InspectionToolList, Name: "列表"},
							},
						},
						{
							ID:   "menu:inspection_item",
							Name: "检测项目",
							Children: []*Node{
								{ID: InspectionItemCreate, Name: "新建"},
								{ID: InspectionItemUpdate, Name: "编辑"},
								{ID: InspectionItemDelete, Name: "删除"},
								{ID: InspectionItemDetail, Name: "详情"},
								{ID: InspectionItemList, Name: "列表"},
							},
						},
					},
				},
				{
					ID:   "menu:defect_management",
					Name: "缺陷管理",
					Children: []*Node{
						{
							ID:   "menu:common_defect",
							Name: "常见缺陷",
							Children: []*Node{
								{ID: CommonDefectCreate, Name: "新建"},
								{ID: CommonDefectUpdate, Name: "编辑"},
								{ID: CommonDefectDelete, Name: "删除"},
								{ID: CommonDefectDetail, Name: "详情"},
								{ID: CommonDefectList, Name: "列表"},
							},
						},
						{
							ID:   "menu:defect_level",
							Name: "缺陷等级",
							Children: []*Node{
								{ID: DefectLevelCreate, Name: "新建"},
								{ID: DefectLevelUpdate, Name: "编辑"},
								{ID: DefectLevelDelete, Name: "删除"},
								{ID: DefectLevelDetail, Name: "详情"},
								{ID: DefectLevelList, Name: "列表"},
							},
						},
						{
							ID:   "menu:defect_location",
							Name: "缺陷位置",
							Children: []*Node{
								{ID: DefectLocationCreate, Name: "新建"},
								{ID: DefectLocationUpdate, Name: "编辑"},
								{ID: DefectLocationDelete, Name: "删除"},
								{ID: DefectLocationDetail, Name: "详情"},
								{ID: DefectLocationList, Name: "列表"},
							},
						},
						{
							ID:   "menu:defect_type",
							Name: "缺陷类型",
							Children: []*Node{
								{ID: DefectTypeCreate, Name: "新建"},
								{ID: DefectTypeUpdate, Name: "编辑"},
								{ID: DefectTypeDelete, Name: "删除"},
								{ID: DefectTypeDetail, Name: "详情"},
								{ID: DefectTypeList, Name: "列表"},
							},
						},
					},
				},
			},
		},
	}

	return permissions, nil
}
