package main //КАК ЖЕ ТУТ ВСЕ ГРЯЗНО

import (
    "github.com/gin-gonic/gin"
    "log"
    "gorm.io/gorm"
    "gorm.io/driver/mysql"
    "github.com/joho/godotenv"
    "fmt"
    "net/http"
    "os"
    "strconv"
    "time"
    "strings"
)


type KPIAlert struct {
    ID         uint      `gorm:"primaryKey" json:"-"`
    ImN        string    `gorm:"column:im_n" json:"im_n"`
    ImSubject  string    `gorm:"column:im_subject" json:"im_subject"`
    Alert      int       `gorm:"column:alert" json:"-"`
    AlertDate  time.Time `gorm:"column:alert_date" json:"alert_date"`
    Done       int       `gorm:"column:done" json:"done"`
    DoneDate   string    `gorm:"column:done_date" json:"-"`
    UpdateDate time.Time `gorm:"column:update_date" json:"-"`
}

type AlertResponse struct {
    ImN       string    `json:"im_n"`
    ImSubject string    `json:"im_subject"`
    Done      int       `json:"done"`
    AlertDate time.Time `json:"alert_date"`
}

func (KPIAlert) TableName() string {
    return "KPI_Alert"
}

func main() {


    if err := godotenv.Load("secretData.env"); err != nil {
        log.Fatal("Ошибка загрузки .env файла")
    }

    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s", os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME"), os.Getenv("DB_PARAMS"))
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal("Ошибка подключения к БД:", err)
    }
    log.Println("Подключено к базе")

    r := gin.Default()

    r.GET("/data", func(c *gin.Context) {
        var alerts []KPIAlert
        var htmlTable strings.Builder

        result := db.Where("done = ?", 0).Find(&alerts)
        if result.Error != nil {
            c.HTML(http.StatusInternalServerError, "<h1>Ошибка! Не удалось получить данные.</h1>", nil)
            return
        }

htmlTable.WriteString(`<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <title>Данные</title>
    <style>
        body {
            display: flex;
            justify-content: center;
            align-items: center;
            min-height: 100vh;
            margin: 0;
            font-family: Arial, sans-serif;
        }
        table {
            border-collapse: collapse;
            width: 80%;
            max-width: 1000px;
            box-shadow: 0 0 10px rgba(0,0,0,0.1);
        }
        th, td {
            border: 1px solid #ddd;
            padding: 8px;
            text-align: left;
        }
        th {
            background-color: #f2f2f2;
        }
        tr {
            cursor: pointer;
        }
        tr:hover {
            background-color: #f5f5f5;
        }
        .jira-link {
            color: #0052cc;
            text-decoration: none;
            display: block; /* Делаем ссылку блочным элементом */
        }
        .jira-link:hover {
            text-decoration: underline;
        }
    </style>
</head>
<body>
    <table>
        <tr>
            <th>Номер задачи</th>
            <th>Описание</th>
            <th>Когда была отправлена</th>
        </tr>`)

for _, alert := range alerts {
    htmlTable.WriteString(fmt.Sprintf(`
    <tr onclick="handleRowClick(event, %d)">
        <td><a class="jira-link" href="https://jira.metro-cc.ru/browse/%s" target="_blank" onclick="event.stopPropagation()">%s</a></td>
        <td>%s</td>
        <td>%s</td>
    </tr>`,
    alert.ID,
    alert.ImN, 
    alert.ImN,
    alert.ImSubject,
    alert.AlertDate.Format(time.RFC3339)))
}

htmlTable.WriteString(`
    </table>
    <script>
        function handleRowClick(event, alertId) {
            // Клик по строке, но не по ссылке
            if (confirm("Вы уверены, что хотите сбросить этот алерт?")) {
                fetch('/data/' + alertId + '/done', {
                    method: 'PUT',
                    headers: {'Content-Type': 'application/json'},
                }).then(response => {
                    if (response.ok) {
                        location.reload();
                    } else {
                        alert("Ошибка при обновлении статуса");
                    }
                });
            }
        }
    </script>
</body>
</html>`)

        c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(htmlTable.String()))
    })

    r.PUT("/data/:id/done", func(c *gin.Context) {
        idStr := c.Param("id")
        id, err := strconv.Atoi(idStr)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
            return
        }

        var kpiAlert KPIAlert
        if err := db.First(&kpiAlert, id).Error; err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "Запись не найдена"})
            return
        }

        newValue := 1 - kpiAlert.Done
        kpiAlert.Done = newValue

        if err := db.Save(&kpiAlert).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сохранения данных"})
            return
        }

        c.Redirect(http.StatusSeeOther, "/data")
    })

    if err := r.Run("10.205.201.13:8080"); err != nil {
        log.Fatal("Ошибка при запуске сервера")
    }
}
