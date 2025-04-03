package workeruc

import (
	"context"
	"encoding/json"
	"github.com/Masterminds/squirrel"
	"github.com/go-ldap/ldap/v3"
	"github.com/tuanta7/qworker/internal/domain"
	"github.com/tuanta7/qworker/pkg/utils"
	"go.uber.org/zap"
	"time"
)

func (u *UseCase) ldapSync(ctx context.Context, connector *domain.Connector, filters ...string) error {
	filter := "(objectClass=*)"
	if len(filters) > 0 {
		filter = filters[0]
	}

	parsedConfig := &domain.LDAPConnector{}
	err := json.Unmarshal(connector.Data.Raw, parsedConfig)
	if err != nil {
		u.logger.Error("ldapSync - json.Unmarshal", zap.Error(err), zap.Any("data", connector.Data.Raw))
		return err
	}

	conn, err := u.ldapClient.NewConnection(parsedConfig.URL, parsedConfig.ConnectTimeout*time.Millisecond)
	if err != nil {
		u.logger.Error("ldapSync - u.ldapClient.NewConnection", zap.Error(err))
		return err
	}
	defer conn.Close()

	pwd, err := u.cipher.DecryptFromStdBase64(parsedConfig.SystemAccountPassword)
	if err != nil {
		u.logger.Error("ldapSync - u.cipher.DecryptFromStdBase64", zap.Error(err))
		return err
	}

	err = conn.Bind(parsedConfig.SystemAccountDN, pwd)
	if err != nil {
		u.logger.Error("ldapSync - conn.Bind", zap.Error(err))
		return err
	}

	pagingControl := ldap.NewControlPaging(parsedConfig.SyncSettings.BatchSize)
	count := 0

	var queries []squirrel.Sqlizer
	for {
		resp, err := conn.Search(&ldap.SearchRequest{
			BaseDN:       parsedConfig.BaseDN,
			TimeLimit:    int(parsedConfig.ReadTimeout),
			SizeLimit:    0,
			Scope:        ldap.ScopeSingleLevel,
			DerefAliases: ldap.NeverDerefAliases,
			Filter:       filter,
			Controls:     []ldap.Control{pagingControl},
		})
		if err != nil {
			u.logger.Error("ldapSync - conn.Search", zap.Error(err))
			return err
		}

		if len(resp.Entries) == 0 {
			break
		}

		count += len(resp.Entries)
		users := make([]*domain.User, len(resp.Entries))
		for i, entry := range resp.Entries {
			user := toUser(entry, connector.Mapper)
			user.SourceID = &connector.ConnectorID
			users[i] = user
		}
		queries = append(queries, u.userRepository.BuildBulkUpsertQuery(users))

		updatedControl := ldap.FindControl(resp.Controls, ldap.ControlTypePaging)
		if ctrl, ok := updatedControl.(*ldap.ControlPaging); ctrl != nil && ok && len(ctrl.Cookie) != 0 {
			pagingControl.SetCookie(ctrl.Cookie)
			continue
		}
		break
	}

	err = u.userRepository.ExecuteTransaction(ctx, queries)
	if err != nil {
		u.logger.Error("ldapSync - u.userRepository.ExecuteTransaction", zap.Error(err))
		return err
	}

	u.logger.Info("sync successfully", zap.Int("count", count))
	return nil
}

func toUser(entry *ldap.Entry, mapper domain.Mapper) *domain.User {
	createdAt, _ := utils.LDAPStringToTime(entry.GetAttributeValue(mapper.CreatedAt))
	updatedAt, _ := utils.LDAPStringToTime(entry.GetAttributeValue(mapper.UpdatedAt))

	return &domain.User{
		FullName:    entry.GetAttributeValue(mapper.FullName),
		Username:    entry.GetAttributeValue(mapper.Username),
		PhoneNumber: entry.GetAttributeValue(mapper.PhoneNumber),
		Email:       entry.GetAttributeValue(mapper.Email),
		Active:      entry.GetAttributeValue(mapper.Custom["active"]) == "9223372036854775807",
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
}
