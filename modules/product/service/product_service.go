package service

import (
	"context"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	"github.com/Caknoooo/go-gin-clean-starter/modules/product/dto"
	"github.com/Caknoooo/go-gin-clean-starter/modules/product/repository"
	pkgdto "github.com/Caknoooo/go-gin-clean-starter/pkg/dto"
	"gorm.io/gorm"
)

type ProductService interface {
	Create(ctx context.Context, req dto.ProductCreateRequest) (dto.ProductResponse, error)
	GetByID(ctx context.Context, id string) (dto.ProductResponse, error)
	List(ctx context.Context, p pkgdto.PaginationRequest) ([]dto.ProductResponse, pkgdto.PaginationResponse, error)
	Update(ctx context.Context, id string, req dto.ProductUpdateRequest) (dto.ProductResponse, error)
	Delete(ctx context.Context, id string) error
}

type productService struct {
	repo repository.ProductRepository
	db   *gorm.DB
}

func NewProductService(repo repository.ProductRepository, db *gorm.DB) ProductService {
	return &productService{repo: repo, db: db}
}

func (s *productService) Create(ctx context.Context, req dto.ProductCreateRequest) (dto.ProductResponse, error) {
	p := entities.Product{
		Name:  req.Name,
		Stock: req.Stock,
	}
	created, err := s.repo.Create(ctx, s.db, p)
	if err != nil {
		return dto.ProductResponse{}, err
	}
	return dto.ProductResponse{ID: created.ID.String(), Name: created.Name, Stock: created.Stock}, nil
}

func (s *productService) GetByID(ctx context.Context, id string) (dto.ProductResponse, error) {
	p, err := s.repo.FindByID(ctx, s.db, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return dto.ProductResponse{}, dto.ErrProductNotFound
		}
		return dto.ProductResponse{}, err
	}
	return dto.ProductResponse{ID: p.ID.String(), Name: p.Name, Stock: p.Stock}, nil
}

func (s *productService) List(ctx context.Context, p pkgdto.PaginationRequest) ([]dto.ProductResponse, pkgdto.PaginationResponse, error) {
	p.Default()
	items, total, err := s.repo.List(ctx, s.db, p.GetLimit(), p.GetOffset())
	if err != nil {
		return nil, pkgdto.PaginationResponse{}, err
	}
	resp := make([]dto.ProductResponse, 0, len(items))
	for _, it := range items {
		resp = append(resp, dto.ProductResponse{ID: it.ID.String(), Name: it.Name, Stock: it.Stock})
	}
	maxPage := total / int64(p.PerPage)
	if total%int64(p.PerPage) != 0 {
		maxPage++
	}
	return resp, pkgdto.PaginationResponse{Page: p.Page, PerPage: p.PerPage, Count: total, MaxPage: maxPage}, nil
}

func (s *productService) Update(ctx context.Context, id string, req dto.ProductUpdateRequest) (dto.ProductResponse, error) {
	p, err := s.repo.FindByID(ctx, s.db, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return dto.ProductResponse{}, dto.ErrProductNotFound
		}
		return dto.ProductResponse{}, err
	}
	if req.Name != "" {
		p.Name = req.Name
	}
	if req.Stock != nil {
		p.Stock = *req.Stock
	}
	updated, err := s.repo.Update(ctx, s.db, p)
	if err != nil {
		return dto.ProductResponse{}, err
	}
	return dto.ProductResponse{ID: updated.ID.String(), Name: updated.Name, Stock: updated.Stock}, nil
}

func (s *productService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, s.db, id)
}
