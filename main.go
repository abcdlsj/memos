//go:generate npm run build
package main

import (
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	//go:embed static
	StaticFS embed.FS

	//go:embed tmpl
	TmplFS embed.FS

	DBFile = orEnv("DB_FILE", "memos.db")
	Port   = orEnv("PORT", "3000")

	DB = initDB(DBFile)
)

type Memo struct {
	ID        int    `json:"id" gorm:"primaryKey"`
	Hero      string `json:"hero" gorm:"type:blob"`
	Tag       string `json:"tag" form:"tag" gorm:"type:text"`
	Title     string `json:"title" form:"title" gorm:"type:text"`
	Content   string `json:"content" form:"content" gorm:"type:text"`
	UpdatedAt int64  `json:"updatedAt" gorm:"autoUpdateTime"`
	CreatedAt int64  `json:"createdAt" gorm:"autoCreateTime"`
	Archived  bool   `json:"archived" gorm:"default:false"`
}

func initDB(filepath string) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(filepath), &gorm.Config{
		DisableAutomaticPing: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	err = db.AutoMigrate(&Memo{})
	if err != nil {
		log.Fatal(err)
	}

	return db
}

func getMemos() ([]Memo, error) {
	var memos []Memo
	if err := DB.Find(&memos, "archived = ?", false).Error; err != nil {
		return nil, err
	}
	return memos, nil
}

func getMemosFilterByTag(tag string) ([]Memo, error) {
	var memos []Memo
	if err := DB.Find(&memos, "tag = ?", tag).Error; err != nil {
		return nil, err
	}
	return memos, nil
}

func getMemo(id int) (*Memo, error) {
	var memo Memo
	if err := DB.First(&memo, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &memo, nil
}

func main() {
	r := gin.Default()

	r.SetHTMLTemplate(template.Must(template.ParseFS(TmplFS, "tmpl/*")))

	staticFS, _ := fs.Sub(StaticFS, "static")

	r.StaticFS("/static/", http.FS(staticFS))

	r.GET("/", func(c *gin.Context) {
		memos, err := getMemos()
		if err != nil {
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{"Message": err.Error()})
			return
		}

		c.HTML(http.StatusOK, "list.html", memos)
	})

	r.GET("/:tag", func(c *gin.Context) {
		memos, err := getMemosFilterByTag(c.Param("tag"))
		if err != nil {
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{"Message": err.Error()})
			return
		}

		c.HTML(http.StatusOK, "list.html", memos)
	})

	r.POST("/memo/create", func(c *gin.Context) {
		var memo Memo

		if err := c.ShouldBind(&memo); err != nil {
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{"Message": err.Error()})
			return
		}
		if err := DB.Create(&memo).Error; err != nil {
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{"Message": err.Error()})
			return
		}

		c.Redirect(http.StatusFound, "/")
	})

	r.GET("/create.html", func(c *gin.Context) {
		c.HTML(http.StatusOK, "create.html", gin.H{})
	})

	r.GET("/timeline.html", func(c *gin.Context) {
		memos, err := getMemos()
		if err != nil {
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{"Message": err.Error()})
			return
		}

		c.HTML(http.StatusOK, "timeline.html", memos)
	})

	r.Run(":" + Port)
}

func orEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}
