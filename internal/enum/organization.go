package enum

import (
	"github.com/xelarion/go-layout/pkg/errs"
)

// OrganizationType 组织类型
type OrganizationType int8

const (
	OrganizationTypeCompany             OrganizationType = iota + 1 // 公司
	OrganizationTypeOffice                                          // 办公室
	OrganizationTypeProjectGroup                                    // 项目组
	OrganizationTypeVirtualOrganization                             // 虚拟组织
	OrganizationTypeDepartment                                      // 部门
)

var OrganizationTypeSlug = map[OrganizationType]string{
	OrganizationTypeCompany:             "company",
	OrganizationTypeOffice:              "office",
	OrganizationTypeProjectGroup:        "project_group",
	OrganizationTypeVirtualOrganization: "virtual_organization",
	OrganizationTypeDepartment:          "department",
}

var OrganizationTypeText = map[OrganizationType]string{
	OrganizationTypeCompany:             "公司",
	OrganizationTypeOffice:              "办公室",
	OrganizationTypeProjectGroup:        "项目组",
	OrganizationTypeVirtualOrganization: "虚拟组织",
	OrganizationTypeDepartment:          "部门",
}

func GetOrganizationType(slug string) (OrganizationType, error) {
	for k, v := range OrganizationTypeSlug {
		if v == slug {
			return k, nil
		}
	}
	return 0, errs.NewBusiness("invalid OrganizationType")
}
