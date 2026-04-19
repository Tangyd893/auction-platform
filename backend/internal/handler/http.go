package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"auction-platform/internal/service"
)

type Handler struct {
	authService  *service.AuthService
	itemService  *service.ItemService
	bidService   *service.BidService
	orderService *service.OrderService
	userService  *service.UserService
}

func NewHandler(auth *service.AuthService, item *service.ItemService, bid *service.BidService, order *service.OrderService, user *service.UserService) *Handler {
	return &Handler{
		authService:  auth,
		itemService:  item,
		bidService:   bid,
		orderService: order,
		userService:  user,
	}
}

// ============ Auth ============

func (h *Handler) Register(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Email    string `json:"email" binding:"required"`
		Role     string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Role == "" {
		req.Role = "buyer"
	}

	user, err := h.authService.Register(req.Username, req.Password, req.Email, req.Role)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, user)
}

func (h *Handler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, user, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token, "user": user})
}

// ============ Items ============

func (h *Handler) CreateItem(c *gin.Context) {
	userID := h.getUserID(c)
	var req struct {
		Title        string `json:"title" binding:"required"`
		Description  string `json:"description"`
		ImageUrl     string `json:"imageUrl"`
		StartPrice   int64  `json:"startPrice" binding:"required"`
		ReservePrice int64  `json:"reservePrice"`
		BidIncrement int64  `json:"bidIncrement"`
		StartTime    int64  `json:"startTime"`
		EndTime      int64  `json:"endTime" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svcReq := &service.CreateItemRequest{
		Title:        req.Title,
		Description:  req.Description,
		ImageUrl:     req.ImageUrl,
		StartPrice:   req.StartPrice,
		ReservePrice: req.ReservePrice,
		BidIncrement: req.BidIncrement,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
	}

	item, err := h.itemService.Create(c.Request.Context(), svcReq, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (h *Handler) GetItem(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	item, err := h.itemService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) ListItems(c *gin.Context) {
	status := c.Query("status")
	keyword := c.Query("keyword")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	items, total, err := h.itemService.List(c.Request.Context(), status, 0, keyword, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": total})
}

func (h *Handler) ListMyItems(c *gin.Context) {
	userID := h.getUserID(c)
	status := c.Query("status")

	items, total, err := h.itemService.ListMyItems(c.Request.Context(), userID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": total})
}

func (h *Handler) CancelItem(c *gin.Context) {
	userID := h.getUserID(c)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	item, err := h.itemService.Cancel(c.Request.Context(), id, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)
}

// ============ Bids ============

func (h *Handler) PlaceBid(c *gin.Context) {
	userID := h.getUserID(c)
	var req struct {
		ItemId int64  `json:"itemId" binding:"required"`
		Amount int64  `json:"amount" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bid, currentPrice, isWinning, err := h.bidService.PlaceBid(c.Request.Context(), req.ItemId, userID, req.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"bidId": bid.ID, "currentPrice": currentPrice, "isWinning": isWinning, "message": "bid placed",
	})
}

func (h *Handler) GetBids(c *gin.Context) {
	itemID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item id"})
		return
	}

	bids, highestPrice, highestBidder, err := h.bidService.GetBidsByItemID(c.Request.Context(), itemID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"bids": bids, "highestPrice": highestPrice, "highestBidderId": highestBidder})
}

func (h *Handler) GetMyBids(c *gin.Context) {
	userID := h.getUserID(c)
	bids, err := h.bidService.GetMyBids(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"bids": bids})
}

// ============ Orders ============

func (h *Handler) CreateOrder(c *gin.Context) {
	userID := h.getUserID(c)
	var req struct {
		ItemId int64 `json:"itemId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := h.orderService.CreateOrder(c.Request.Context(), req.ItemId, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, order)
}

func (h *Handler) GetOrder(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	order, err := h.orderService.GetOrder(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}
	c.JSON(http.StatusOK, order)
}

func (h *Handler) ListOrders(c *gin.Context) {
	userID := h.getUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	orders, total, err := h.orderService.ListOrders(c.Request.Context(), userID, "", page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"orders": orders, "total": total})
}

func (h *Handler) UpdateOrderStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := h.orderService.UpdateOrderStatus(c.Request.Context(), id, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, order)
}

// ============ Helper ============

func (h *Handler) getUserID(c *gin.Context) int64 {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return 1 // 测试用，未登录默认用户1
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// 解析 JWT（不验证签名，仅提取 userID）
	// 生产环境应验证签名
	token, _, err := jwt.NewParser().ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return 1
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if userID, ok := claims["user_id"].(float64); ok {
			return int64(userID)
		}
	}
	return 1
}
