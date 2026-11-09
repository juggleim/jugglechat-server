package dbcommons

import (
	"bufio"
	"embed"
	"errors"
	"fmt"
	mysqlerr "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"sort"
	"strings"

	utils "github.com/juggleim/jugglechat-server/commons/tools"
)

//go:embed sqls/*
var sqlFs embed.FS

const (
	JChatDbVersionKey = "jchatdb_version"
)

type CountResult struct {
	Count int64 `gorm:"count"`
}

func Upgrade() error {
	// upgrade jchat db
	var currVersion int64 = 0
	dao := GlobalConfDao{}
	conf, err := dao.FindByKey(JChatDbVersionKey)
	if err == nil && conf != nil {
		ver, err := utils.String2Int64(conf.ConfValue)
		if err == nil && ver > 0 {
			currVersion = ver
		}
	}
	fmt.Println("[JChatDbMigration]current version:", currVersion)
	sqlFiles, err := sqlFs.ReadDir("sqls")
	if err == nil {
		neededVers := []int64{}
		for _, sqlFile := range sqlFiles {
			fileName := sqlFile.Name()
			if len(fileName) == 12 {
				fileName = fileName[:8]
			}
			ver, err := utils.String2Int64(fileName)
			if err == nil && ver > 0 {
				neededVers = append(neededVers, ver)
			}
		}
		//sort
		sort.Slice(neededVers, func(i, j int) bool {
			return neededVers[i] < neededVers[j]
		})
		for _, ver := range neededVers {
			if ver > currVersion {
				sqlFileName := fmt.Sprintf("sqls/%d.sql", ver)
				fmt.Println("[DbMigration]start to execute sql file:", sqlFileName)
				err := executeSqlFile(sqlFileName)
				if err == nil {
					fmt.Println("[DbMigration]execute sql file success:", sqlFileName)
					if err := dao.Upsert(GlobalConfDao{
						ConfKey:   JChatDbVersionKey,
						ConfValue: fmt.Sprintf("%d", ver),
					}); err != nil {
						return err
					}
				} else {
					return err
				}
			}
		}
	}
	return err
}

func executeSqlFile(fileName string) error {
	sqlFile, err := sqlFs.Open(fileName)
	if err != nil {
		fmt.Println("[DbMigration_Err]Read sql file err:", err, "file_name:", fileName)
		return err
	}
	defer sqlFile.Close()

	b, err := sqlFs.ReadFile(fileName)
	if err != nil {
		return err
	}
	// Some idempotent migrations use session-scoped PREPARE/user variables.
	return GetDb().Connection(func(conn *gorm.DB) error {
		if err := runSQLStatements(string(b), func(query string) error { return conn.Exec(query).Error }); err != nil {
			return err
		}
		if fileName == "sqls/20261109.sql" {
			return EnsurePasswordSMSPhoneTrigger(conn)
		}
		return nil
	})
}

// CREATE TRIGGER is not preparable in MySQL. Check metadata and execute a
// single ordinary DDL statement, without 8.0.29's IF NOT EXISTS syntax.
func EnsurePasswordSMSPhoneTrigger(conn *gorm.DB) error {
	const name = "users_password_sms_phone_version"
	var count int64
	if err := conn.Raw("SELECT COUNT(*) FROM information_schema.TRIGGERS WHERE TRIGGER_SCHEMA=DATABASE() AND TRIGGER_NAME=?", name).Scan(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	err := conn.Exec("CREATE TRIGGER users_password_sms_phone_version BEFORE UPDATE ON users FOR EACH ROW SET NEW.password_sms_phone_version = OLD.password_sms_phone_version + IF(CAST(NEW.phone AS BINARY) <=> CAST(OLD.phone AS BINARY),0,1)").Error
	if err == nil {
		return nil
	}
	// Another startup may have installed it between metadata read and DDL.
	var me *mysqlerr.MySQLError
	if errors.As(err, &me) && me.Number == 1359 {
		return nil
	}
	return err
}

func runSQLStatements(content string, execute func(string) error) error {
	scanner := bufio.NewScanner(strings.NewReader(content))
	var queryBuilder strings.Builder
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "--") {
			continue
		}
		queryBuilder.WriteString(line)
		queryBuilder.WriteByte(' ')
		if strings.HasSuffix(line, ";") {
			query := strings.TrimSpace(queryBuilder.String())
			if query != "" {
				if err := execute(query); err != nil {
					return err
				}
			}
			queryBuilder.Reset()
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if strings.TrimSpace(queryBuilder.String()) != "" {
		return fmt.Errorf("unterminated migration statement")
	}
	return nil
}
