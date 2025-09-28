# 项目与文件结构命名规范

* 项目名 (Project Name)：使用小写的 kebab-case (短横线连接)
* 包名 (Package Name)：使用简短、小写、有意义的单个单词
* 文件夹/目录名 (Folder/Directory Name)：与包名保持一致，全小写
* 文件名 (File Name)：使用小写的 snake_case.go

# 代码元素命名规范

* 变量名 (Variable Name)：使用驼峰命名法 (camelCase)
* 常量名 (Constant Name)：与变量名一样，使用驼峰命名法 (CamelCase 或 camelCase)
* 函数/方法名 (Function/Method Name)：使用驼峰命名法 (CamelCase)
* 接口名 (Interface Name)：
  * 单个方法的接口： 接口名以 "er" 结尾。这是最地道 (Idiomatic) 的方式;
  * 多个方法的接口： 接口名通常是一个名词，描述其代表的角色或能力。
* 结构体名 (Struct Name)：使用驼峰命名法 (CamelCase)

# 注意

* 文件名：[领域/实体]_[实现细节].go 
* 结构体名：[细节][领域/实体] eg：
* 缩写词要么全大写，要么全小写 eg：JWT、URL、HTTP、ID
* 结构体成员：[主体名词][类型缩写] eg：userSvc
* 函数和方法名：[动作][上下文] or [动作][上下文]