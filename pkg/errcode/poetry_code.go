package errcode

var (
	// 诗词模块错误 (3000 ~ 3999)
	DynastyHasAuthors = NewError(3001, "该朝代下仍有诗人，禁止删除")
	GenreHasWorks     = NewError(3002, "该体裁下仍有作品，禁止删除")
	AuthorHasWorks    = NewError(3003, "该诗人名下仍有作品，禁止删除")

	// 唯一性约束
	DynastyNameExist = NewError(3004, "朝代名称已存在")
	GenreNameExist   = NewError(3005, "体裁名称已存在")
)
