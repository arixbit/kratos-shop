package data

import (
	"context"
	"encoding/json"
	"goods/internal/biz"
	"goods/internal/domain"
	"strconv"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/olivere/elastic/v7"
)

// GetIndexName 设计商品的索引 goods
func (esGoodsRepo) GetIndexName() string {
	return "goods"
}

// GetMapping 设计商品的 mapping 结构
func (esGoodsRepo) GetMapping() string {
	goodsMapping := `
	{
    "mappings": {
        "properties": {
            "id": {
                "type": "integer"
            },
            "brands_id": {
                "type": "integer"
            },
            "category_id": {
                "type": "integer"
            },
            "type_id": {
                "type": "integer"
            },
            "click_num": {
                "type": "integer"
            },
            "fav_num": {
                "type": "integer"
            },
            "is_hot": {
                "type": "boolean"
            },
            "is_new": {
                "type": "boolean"
            },
            "market_price": {
                "type": "integer"
            },
            "name": {
                "type": "text",
                "analyzer": "standard"
            },
			"brand_name": {
                "type": "keyword",
                "index": false
            },
			"category_name": {
                "type": "keyword",
                "index": false
            },
			"type_name": {
                "type": "keyword",
                "index": false
            },
            "goods_brief": {
                "type": "text",
                "analyzer": "standard"
            },
            "on_sale": {
                "type": "boolean"
            },
            "ship_free": {
                "type": "boolean"
            },
            "shop_price": {
                "type": "integer"
            },
            "sold_num": {
                "type": "integer"
            },
			"sku": {
				"type": "nested",
				"properties": {
					"sku_id": {
						"type": "integer"
					},
					"sku_name": {
						"type": "text",
						"analyzer": "standard"
					},
					"sku_price": {
						"type": "integer"
					}
				}
			}
        }
    }
}`
	return goodsMapping
}

type esGoodsRepo struct {
	data *Data
	log  *log.Helper
}

// NewEsGoodsRepo .
func NewEsGoodsRepo(data *Data, logger log.Logger) biz.EsGoodsRepo {
	return &esGoodsRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (p esGoodsRepo) GoodsList(ctx context.Context, filter *domain.EsSearch) ([]int64, int64, error) {
	if err := p.ensureIndex(ctx); err != nil {
		return nil, 0, err
	}
	boolQuery := elastic.NewBoolQuery()
	boolQuery.Must(filter.MustQuery...)
	boolQuery.MustNot(filter.MustNotQuery...)
	boolQuery.Should(filter.ShouldQuery...)
	boolQuery.Filter(filter.Filters...)

	result, err := p.data.EsClient.Search().
		Index(p.GetIndexName()).
		Query(boolQuery).
		SortBy(filter.Sorters...).
		From(int(filter.Form)).
		Size(int(filter.Size)).
		Do(ctx)

	if err != nil {
		return nil, 0, err
	}

	// 取出来商品ID
	goodsIds := make([]int64, 0)
	for _, value := range result.Hits.Hits {
		goods := domain.ESGoods{}
		_ = json.Unmarshal(value.Source, &goods)
		goodsIds = append(goodsIds, goods.ID)
	}
	return goodsIds, result.Hits.TotalHits.Value, nil

}

func (p esGoodsRepo) InsertEsGoods(ctx context.Context, esModel domain.ESGoods) error {
	if err := p.ensureIndex(ctx); err != nil {
		return err
	}
	_, err := p.data.EsClient.Index().Index(p.GetIndexName()).BodyJson(esModel).Id(strconv.Itoa(int(esModel.ID))).Do(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (p esGoodsRepo) ensureIndex(ctx context.Context) error {
	exists, err := p.data.EsClient.IndexExists(p.GetIndexName()).Do(ctx)
	if err != nil {
		return err
	}
	if !exists {
		_, err = p.data.EsClient.CreateIndex(p.GetIndexName()).BodyString(p.GetMapping()).Do(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}
