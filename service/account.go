package service

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"h-ui/dao"
	"h-ui/model/bo"
	"h-ui/model/constant"
	"h-ui/model/dto"
	"h-ui/model/entity"
	"h-ui/model/vo"
)

func Login(username string, pass string) (string, error) {
	account, err := dao.GetAccount("username = ? and pass = ? and role = 'admin' and deleted = 0", username, pass)
	if err != nil {
		return "", err
	}
	accountBo := bo.AccountBo{
		Id:       *account.Id,
		Username: *account.Username,
		Roles:    []string{*account.Role},
		Deleted:  *account.Deleted,
	}
	return GenToken(accountBo)
}

func PageAccount(accountPageDto dto.AccountPageDto) ([]entity.Account, int64, error) {
	return dao.PageAccount(accountPageDto)
}

func SaveAccount(account entity.Account) error {
	_, err := dao.SaveAccount(account)
	return err
}

func DeleteAccount(ids []int64) error {
	return dao.DeleteAccount(ids)
}

func UpdateAccount(account entity.Account) error {
	updates := map[string]interface{}{}
	if account.Username != nil && *account.Username != "" {
		updates["username"] = *account.Username
	}
	if account.Pass != nil && *account.Pass != "" {
		updates["pass"] = *account.Pass
	}
	if account.ConPass != nil && *account.ConPass != "" {
		updates["con_pass"] = fmt.Sprintf("%s.%s", *account.Username, *account.ConPass)
	}
	if account.Quota != nil {
		updates["quota"] = *account.Quota
	}
	if account.ExpireTime != nil {
		updates["expire_time"] = *account.ExpireTime
	}
	if account.Download != nil {
		updates["download"] = *account.Download
	}
	if account.Upload != nil {
		updates["upload"] = *account.Upload
	}
	if account.DeviceNo != nil {
		updates["device_no"] = *account.DeviceNo
	}
	if account.NodeAccess != nil {
		updates["node_access"] = *account.NodeAccess
	}
	if account.Deleted != nil {
		updates["deleted"] = *account.Deleted
	}
	if account.LoginAt != nil && *account.LoginAt > 0 {
		updates["login_at"] = *account.LoginAt
	}
	if account.ConAt != nil && *account.ConAt > 0 {
		updates["con_at"] = *account.ConAt
	}
	return dao.UpdateAccount([]int64{*account.Id}, updates)
}

func ResetTraffic(id int64) error {
	return dao.UpdateAccount([]int64{id}, map[string]interface{}{"download": 0, "upload": 0})
}

func ExistAccountUsername(username string, id int64) bool {
	var err error
	if id != 0 {
		_, err = dao.GetAccount("username = ? and id != ?", username, id)
	} else {
		_, err = dao.GetAccount("username = ?", username)
	}
	if err != nil {
		if err.Error() == constant.WrongPassword {
			return false
		}
	}
	return true
}

func GetAccount(id int64) (entity.Account, error) {
	return dao.GetAccount("id = ?", id)
}

func ListExportAccount() ([]bo.AccountExport, error) {
	accounts, err := dao.ListAccount(nil, nil)
	if err != nil {
		return nil, errors.New(constant.SysError)
	}
	var accountExports []bo.AccountExport
	for _, item := range accounts {
		accountExport := bo.AccountExport{
			Id:           *item.Id,
			Username:     *item.Username,
			Pass:         *item.Pass,
			ConPass:      *item.ConPass,
			Quota:        *item.Quota,
			Download:     *item.Download,
			Upload:       *item.Upload,
			ExpireTime:   *item.ExpireTime,
			DeviceNo:     *item.DeviceNo,
			KickUtilTime: *item.KickUtilTime,
			Role:         *item.Role,
			NodeAccess:   *item.NodeAccess,
			Deleted:      *item.Deleted,
			CreateTime:   *item.CreateTime,
			UpdateTime:   *item.UpdateTime,
			LoginAt:      *item.LoginAt,
			ConAt:        *item.ConAt,
		}
		accountExports = append(accountExports, accountExport)
	}
	return accountExports, nil
}

func ReleaseKickAccount(id int64) error {
	return dao.UpdateAccount([]int64{id}, map[string]interface{}{"kick_util_time": 0})
}

func UpsertAccount(accounts []entity.Account) error {
	return dao.UpsertAccount(accounts)
}

func GetAccountInfo(c *gin.Context) (vo.AccountInfoVo, error) {
	myClaims, err := ParseToken(GetToken(c))
	if err != nil {
		return vo.AccountInfoVo{}, err
	}
	if myClaims.AccountBo.Deleted != 0 {
		return vo.AccountInfoVo{}, errors.New("this account has been disabled")
	}
	return vo.AccountInfoVo{
		Id:       myClaims.AccountBo.Id,
		Username: myClaims.AccountBo.Username,
		Roles:    myClaims.AccountBo.Roles,
	}, nil
}
// ValidateNodeAccess 验证用户节点权限
func ValidateNodeAccess(nodeAccess int64) error {
	if nodeAccess != 1 && nodeAccess != 2 {
		return errors.New("invalid node access value, must be 1 or 2")
	}
	
	// 如果用户要求双节点权限，检查第二节点是否启用
	if nodeAccess == 2 {
		enabled, err := IsNode2Enabled()
		if err != nil {
			return err
		}
		if !enabled {
			return errors.New("node2 is not enabled, cannot set dual node access")
		}
	}
	
	return nil
}

// GetUserNodeAccess 获取用户节点权限
func GetUserNodeAccess(accountId int64) (int64, error) {
	account, err := GetAccount(accountId)
	if err != nil {
		return 1, err
	}
	
	if account.NodeAccess == nil {
		return 1, nil // 默认单节点
	}
	
	return *account.NodeAccess, nil
}

// SetDefaultNodeAccess 为现有用户设置默认节点权限
func SetDefaultNodeAccess() error {
	// 这个函数在数据库迁移时调用，为所有现有用户设置默认的单节点权限
	accounts, err := dao.ListAccount(nil, nil)
	if err != nil {
		return err
	}
	
	for _, account := range accounts {
		if account.NodeAccess == nil {
			defaultAccess := int64(1)
			updates := map[string]interface{}{"node_access": defaultAccess}
			if err := dao.UpdateAccount([]int64{*account.Id}, updates); err != nil {
				return err
			}
		}
	}
	
	return nil
}// 
DowngradeUsersToSingleNode 将所有双节点用户降级为单节点
func DowngradeUsersToSingleNode() error {
	// 查找所有双节点权限的用户
	accounts, err := dao.ListAccount("node_access = ?", []interface{}{2})
	if err != nil {
		return err
	}

	if len(accounts) == 0 {
		return nil // 没有需要降级的用户
	}

	// 批量更新为单节点权限
	var userIds []int64
	for _, account := range accounts {
		userIds = append(userIds, *account.Id)
	}

	updates := map[string]interface{}{"node_access": 1}
	return dao.UpdateAccount(userIds, updates)
}