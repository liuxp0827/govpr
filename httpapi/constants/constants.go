package constants

const (
	SUCCESS_REGISTER_USER   = 1001 // 注册用户成功
	SUCCESS_DELETE_USER     = 1002 // 删除用成功
	SUCCESS_CLEAR_SAMPLES   = 1003 // 清除数据库中训练语音缓存成功
	SUCCESS_TRAIN_MODEL     = 1004 // 训练模型成功
	SUCCESS_DELETE_MODEL    = 1005 // 删除模型成功
	SUCCESS_ADDSAMPLE       = 1006 // 添加语音到数据库训练语音缓存成功
	SUCCESS_VERIFY_MODEL    = 1007 // 验证成功
	SUCCESS_DETECT_REGISTER = 1010 // 登记检测通过
	SUCCESS_DETECT_QUERY    = 1011 // 验证检测通过

	FAILED_REGISTER_USER   = 100 // 注册用户失败
	FAILED_DELETE_USER     = 200 // 删除用失败
	FAILED_CLEAR_SAMPLES   = 300 // 清除数据库中训练语音缓存失败
	FAILED_TRAIN_MODEL     = 400 // 训练模型失败
	FAILED_DELETE_MODEL    = 500 // 删除模型失败
	FAILED_ADDSAMPLE       = 600 // 添加语音到数据库训练语音缓存失败
	FAILED_VERIFY_MODEL    = 700 // 验证失败
	FAILED_DETECT_REGISTER = 800 // 登记检测失败
	FAILED_DETECT_QUERY    = 900 // 验证检测失败

	ERROR_USER_EXISTENT        = 2001 // 用户已存在
	ERROR_USER_NONEXISTENT     = 2002 // 用户不存在
	ERROR_CLEAR_SAMPLES_FAILED = 2003 // 清除数据库中训练语音缓存失败
	ERROR_MODEL_EXISTENT       = 2004 // 模型已存在
	ERROR_MODEL_NONEXISTENT    = 2005 // 模型不存在
	ERROR_SAMPLES_NOT_ENOUGH   = 2006 // 数据库中训练语音缓存条数不足
	ERROR_LENGTH_MISMATCH      = 2007 // 数据库中训练语音缓存条数与数据库中文本内容条数不同
	ERROR_TRAIN_MODEL_FAILED   = 2008 // 训练模型失败
	ERROR_SAMPLE_IS_NULL       = 2009 // 上传语音为空
	ERROR_ADDSAMPLE_FAILED     = 2010 // 添加语音到数据库训练语音缓存失败
	ERROR_VERIFY_MODEL_FAILED  = 2011 // 验证失败
	ERROR_USER_ILLEGAL         = 2012 // 用户名不合法
	ERROR_APP_TOKEN            = 2018 // 权限不合法
	ERROR_URL_PARAM_ILLEGAL    = 2019 // url参数不合法
)
