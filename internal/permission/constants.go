package permission

// Permission constants
const (
	// All permissions
	All = "*"

	RoleList          = "role/list"    // Role list
	RoleDetail        = "role/detail"  // Role detail
	RoleCreate        = "role/create"  // Create role
	RoleUpdate        = "role/update"  // Update role
	RoleUpdateEnabled = "role/enabled" // Update role enabled
	RoleDelete        = "role/delete"  // Delete role

	OrganizationList   = "organization/list"   // Organization list
	OrganizationCreate = "organization/create" // Create organization
	OrganizationUpdate = "organization/update" // Update organization
	OrganizationDelete = "organization/delete" // Delete organization

	PositionList   = "position/list"   // Position list
	PositionCreate = "position/create" // Create position
	PositionUpdate = "position/update" // Update position
	PositionDelete = "position/delete" // Delete position

	PermissionUpdate = "permission/update" // Update permission

	CodeRuleList   = "code_rule/list"   // Code rule list
	CodeRuleUpdate = "code_rule/update" // Update code rule

	UserList           = "user/list"     // User list
	UserCreate         = "user/create"   // Create user
	UserUpdate         = "user/update"   // Update user
	UserUpdateEnabled  = "user/enabled"  // Update user enabled
	UserDelete         = "user/delete"   // Delete user
	UserExport         = "user/export"   // Export user
	UserUpdatePassword = "user/password" // Update user password

	BatchRecordList   = "batch_record/list"   // Batch record list
	BatchRecordDetail = "batch_record/detail" // Batch record detail

	MaterialList          = "material/list"    // 列表
	MaterialDetail        = "material/detail"  // 详情
	MaterialCreate        = "material/create"  // 创建
	MaterialUpdate        = "material/update"  // 更新
	MaterialUpdateEnabled = "material/enabled" // 更新状态
	MaterialDelete        = "material/delete"  // 删除
	MaterialExport        = "material/export"  // 导出

	AnnouncementList   = "announcement/list"   // 列表
	AnnouncementDetail = "announcement/detail" // 详情
	AnnouncementCreate = "announcement/create" // 创建
	AnnouncementUpdate = "announcement/update" // 更新
	AnnouncementDelete = "announcement/delete" // 删除

	CustomerList   = "customer/list"   // Customer list
	CustomerDetail = "customer/detail" // Customer detail
	CustomerCreate = "customer/create" // Create customer
	CustomerUpdate = "customer/update" // Update customer
	CustomerDelete = "customer/delete" // Delete customer
	CustomerImport = "customer/import" // Import customer
	CustomerExport = "customer/export"

	SupplierList   = "supplier/list"   // Supplier list
	SupplierDetail = "supplier/detail" // Supplier detail
	SupplierCreate = "supplier/create" // Create supplier
	SupplierUpdate = "supplier/update" // Update supplier
	SupplierDelete = "supplier/delete" // Delete supplier
	SupplierImport = "supplier/import" // Import customer
	SupplierExport = "supplier/export"

	UnitList   = "unit/list"   // User list
	UnitDetail = "unit/detail" // User detail
	UnitCreate = "unit/create" // Create user
	UnitUpdate = "unit/update" // Update user
	UnitDelete = "unit/delete" // Delete user

	EquipmentList   = "equipment/list"
	EquipmentDetail = "equipment/detail"
	EquipmentCreate = "equipment/create"
	EquipmentUpdate = "equipment/update"
	EquipmentDelete = "equipment/delete"
	EquipmentExport = "equipment/export"
	EquipmentImport = "equipment/import"

	EquipmentTypeList   = "equipment_type/list"   // EquipmentType list
	EquipmentTypeDetail = "equipment_type/detail" // EquipmentType detail
	EquipmentTypeCreate = "equipment_type/create" // Create equipment_type
	EquipmentTypeUpdate = "equipment_type/update" // Update equipment_type
	EquipmentTypeDelete = "equipment_type/delete" // Delete equipment_type
	EquipmentTypeExport = "equipment_type/export" // Export equipment_type
	EquipmentTypeImport = "equipment_type/import" // Import equipment_type

	WarehouseList   = "warehouse/list"   // Warehouse list
	WarehouseDetail = "warehouse/detail" // Warehouse detail
	WarehouseCreate = "warehouse/create" // Create warehouse
	WarehouseUpdate = "warehouse/update" // Update warehouse
	WarehouseDelete = "warehouse/delete" // Delete warehouse

	WarehouseLocationList   = "warehouse_location/list"   // warehouseLocation list
	WarehouseLocationDetail = "warehouse_location/detail" // warehouseLocation detail
	WarehouseLocationCreate = "warehouse_location/create" // Create warehouse_location
	WarehouseLocationUpdate = "warehouse_location/update" // Update warehouse_location
	WarehouseLocationDelete = "warehouse_location/delete" // Delete warehouse_location

	WarehouseAreaList   = "warehouse_area/list"   // warehouseArea list
	WarehouseAreaDetail = "warehouse_area/detail" // warehouseArea detail
	WarehouseAreaCreate = "warehouse_area/create" // Create warehouse_area
	WarehouseAreaUpdate = "warehouse_area/update" // Update warehouse_area
	WarehouseAreaDelete = "warehouse_area/delete" // Delete warehouse_area

	InspectionItemList   = "inspection_item/list"
	InspectionItemDetail = "inspection_item/detail"
	InspectionItemCreate = "inspection_item/create"
	InspectionItemUpdate = "inspection_item/update"
	InspectionItemDelete = "inspection_item/delete"

	InspectionToolList   = "inspection_tool/list"
	InspectionToolDetail = "inspection_tool/detail"
	InspectionToolCreate = "inspection_tool/create"
	InspectionToolUpdate = "inspection_tool/update"
	InspectionToolDelete = "inspection_tool/delete"

	DefectLevelList   = "defect_level/list"
	DefectLevelDetail = "defect_level/detail"
	DefectLevelCreate = "defect_level/create"
	DefectLevelUpdate = "defect_level/update"
	DefectLevelDelete = "defect_level/delete"

	DefectTypeList   = "defect_type/list"
	DefectTypeDetail = "defect_type/detail"
	DefectTypeCreate = "defect_type/create"
	DefectTypeUpdate = "defect_type/update"
	DefectTypeDelete = "defect_type/delete"

	DefectLocationList   = "defect_location/list"
	DefectLocationDetail = "defect_location/detail"
	DefectLocationCreate = "defect_location/create"
	DefectLocationUpdate = "defect_location/update"
	DefectLocationDelete = "defect_location/delete"

	CommonDefectList   = "common_defect/list"
	CommonDefectDetail = "common_defect/detail"
	CommonDefectCreate = "common_defect/create"
	CommonDefectUpdate = "common_defect/update"
	CommonDefectDelete = "common_defect/delete"

	ProductList          = "product/list"    // 列表
	ProductDetail        = "product/detail"  // 详情
	ProductCreate        = "product/create"  // 创建
	ProductUpdate        = "product/update"  // 更新
	ProductUpdateEnabled = "product/enabled" // 更新状态
	ProductDelete        = "product/delete"  // 删除
	ProductExport        = "product/export"  // 导出

	BomList   = "bom/list"   // 列表
	BomDetail = "bom/detail" // 详情
	BomCreate = "bom/create" // 创建
	BomUpdate = "bom/update" // 更新
	BomDelete = "bom/delete" // 删除

	ProcessList   = "process/list"   // 列表
	ProcessCreate = "process/create" // 创建
	ProcessUpdate = "process/update" // 更新
	ProcessDelete = "process/delete" // 删除
	ProcessExport = "process/export" // 导出

	ProcessRouteList   = "process_route/list"   // 列表
	ProcessRouteDetail = "process_route/detail" // 详情
	ProcessRouteCreate = "process_route/create" // 创建
	ProcessRouteUpdate = "process_route/update" // 更新
	ProcessRouteDelete = "process_route/delete" // 删除
	ProcessRouteExport = "process_route/export" // 导出

	ProductionPlanList   = "production_plan/list"   // 列表
	ProductionPlanDetail = "production_plan/detail" // 详情
	ProductionPlanCreate = "production_plan/create" // 创建
	ProductionPlanUpdate = "production_plan/update" // 更新
	ProductionPlanDelete = "production_plan/delete" // 删除
	ProductionPlanExport = "production_plan/export" // 导出
)
