package entity

import "errors"

// 地址相关错误
var (
	ErrAddressNotFound       = errors.New("地址不存在")
	ErrAddressAlreadyDefault = errors.New("地址已是默认地址")
	ErrCannotDeleteDefault   = errors.New("不能删除默认地址")
	ErrInvalidAddressID      = errors.New("无效的地址ID")
	ErrRecipientNameRequired = errors.New("收货人姓名不能为空")
	ErrPhoneRequired         = errors.New("联系电话不能为空")
	ErrProvinceRequired      = errors.New("省份不能为空")
	ErrCityRequired          = errors.New("城市不能为空")
	ErrDistrictRequired      = errors.New("区县不能为空")
	ErrDetailRequired        = errors.New("详细地址不能为空")
)

// 用户相关错误
var (
	ErrUserNotFound        = errors.New("用户不存在")
	ErrUserAlreadyExists   = errors.New("用户已存在")
	ErrInvalidPassword     = errors.New("密码错误")
	ErrInvalidUserID       = errors.New("无效的用户ID")
	ErrInvalidUsername     = errors.New("无效的用户名")
	ErrInvalidEmail        = errors.New("无效的邮箱地址")
	ErrEmailAlreadyExists  = errors.New("邮箱已存在")
)

// 商品相关错误
var (
	ErrProductNotFound      = errors.New("商品不存在")
	ErrProductOutOfStock    = errors.New("商品库存不足")
	ErrProductNotAvailable  = errors.New("商品已下架")
	ErrProductAlreadyExists = errors.New("商品已存在")
	ErrInvalidProductID     = errors.New("无效的商品ID")
	ErrInvalidCategoryID    = errors.New("无效的分类ID")
	ErrInvalidStock         = errors.New("无效的库存数量")
	ErrInvalidPrice         = errors.New("无效的商品价格")
)

// 分类相关错误
var (
	ErrCategoryNotFound      = errors.New("分类不存在")
	ErrCategoryAlreadyExists = errors.New("分类已存在")
)

// 订单相关错误
var (
	ErrOrderNotFound       = errors.New("订单不存在")
	ErrOrderStatusInvalid  = errors.New("订单状态无效")
	ErrOrderCannotCancel   = errors.New("订单无法取消")
	ErrOrderAlreadyPaid    = errors.New("订单已支付")
)

// 购物车相关错误
var (
	ErrCartItemNotFound = errors.New("购物车商品不存在")
	ErrCartEmpty        = errors.New("购物车为空")
)

// 退款相关错误
var (
	ErrRefundNotFound        = errors.New("退款申请不存在")
	ErrRefundAlreadyExists   = errors.New("该订单已存在退款申请")
	ErrInvalidRefundID       = errors.New("无效的退款ID")
	ErrInvalidRefundType     = errors.New("无效的退款类型")
	ErrInvalidRefundStatus   = errors.New("无效的退款状态")
	ErrInvalidRefundAmount   = errors.New("无效的退款金额")
	ErrRefundCannotApprove   = errors.New("退款申请无法审核通过")
	ErrRefundCannotReject    = errors.New("退款申请无法审核拒绝")
	ErrRefundCannotComplete  = errors.New("退款申请无法完成")
	ErrRefundReasonRequired  = errors.New("退款原因不能为空")
)

// 评价相关错误
var (
	ErrReviewNotFound        = errors.New("评价不存在")
	ErrReviewAlreadyExists   = errors.New("该订单和商品已存在评价")
	ErrInvalidReviewID       = errors.New("无效的评价ID")
	ErrInvalidRating         = errors.New("无效的评分")
	ErrReviewContentRequired = errors.New("评价内容不能为空")
)
