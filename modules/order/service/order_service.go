package service

import (
    "context"

    "github.com/Caknoooo/go-gin-clean-starter/database/entities"
    "github.com/Caknoooo/go-gin-clean-starter/modules/order/dto"
    "github.com/Caknoooo/go-gin-clean-starter/modules/order/repository"
    productRepo "github.com/Caknoooo/go-gin-clean-starter/modules/product/repository"
    pkgdto "github.com/Caknoooo/go-gin-clean-starter/pkg/dto"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

type OrderService interface {
    Create(ctx context.Context, req dto.OrderCreateRequest) (dto.OrderResponse, error)
    GetByID(ctx context.Context, id string) (dto.OrderResponse, error)
    List(ctx context.Context, p pkgdto.PaginationRequest) ([]dto.OrderResponse, pkgdto.PaginationResponse, error)
    Delete(ctx context.Context, id string) error
}

type orderService struct {
    orderRepository   repository.OrderRepository
    productRepository productRepo.ProductRepository
    db                *gorm.DB
}

func NewOrderService(orderRepo repository.OrderRepository, prodRepo productRepo.ProductRepository, db *gorm.DB) OrderService {
    return &orderService{orderRepository: orderRepo, productRepository: prodRepo, db: db}
}

func (s *orderService) Create(ctx context.Context, req dto.OrderCreateRequest) (dto.OrderResponse, error) {
    pid, err := uuid.Parse(req.ProductID)
    if err != nil {
        return dto.OrderResponse{}, err
    }

	var created entities.Order
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Ensure product exists and has sufficient stock
		ok, err := s.productRepository.DecrementStock(ctx, tx, pid, req.Quantity)
		if err != nil {
			return err
		}
		if !ok {
			return dto.ErrInsufficientStock
		}

		order := entities.Order{
			ProductID: pid,
			BuyerID:   req.BuyerID,
			Quantity:  req.Quantity,
		}

		var errCreate error
		created, errCreate = s.orderRepository.Create(ctx, tx, order)
		return errCreate
	})
	if err != nil {
		return dto.OrderResponse{}, err
	}

	return dto.OrderResponse{
		ID:        created.ID.String(),
		ProductID: created.ProductID.String(),
		BuyerID:   created.BuyerID,
		Quantity:  created.Quantity,
	}, nil
}

func (s *orderService) GetByID(ctx context.Context, id string) (dto.OrderResponse, error) {
	o, err := s.orderRepository.FindByID(ctx, s.db, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return dto.OrderResponse{}, dto.ErrOrderNotFound
		}
		return dto.OrderResponse{}, err
	}
	return dto.OrderResponse{
		ID:        o.ID.String(),
		ProductID: o.ProductID.String(),
		BuyerID:   o.BuyerID,
		Quantity:  o.Quantity,
	}, nil
}

func (s *orderService) List(ctx context.Context, p pkgdto.PaginationRequest) ([]dto.OrderResponse, pkgdto.PaginationResponse, error) {
	p.Default()
	items, total, err := s.orderRepository.List(ctx, s.db, p.GetLimit(), p.GetOffset())
	if err != nil {
		return nil, pkgdto.PaginationResponse{}, err
	}
	resp := make([]dto.OrderResponse, 0, len(items))
	for _, it := range items {
		resp = append(resp, dto.OrderResponse{
			ID:        it.ID.String(),
			ProductID: it.ProductID.String(),
			BuyerID:   it.BuyerID,
			Quantity:  it.Quantity,
		})
	}
	maxPage := total / int64(p.PerPage)
	if total%int64(p.PerPage) != 0 {
		maxPage++
	}
	return resp, pkgdto.PaginationResponse{Page: p.Page, PerPage: p.PerPage, Count: total, MaxPage: maxPage}, nil
}

func (s *orderService) Delete(ctx context.Context, id string) error {
	return s.orderRepository.Delete(ctx, s.db, id)
}
