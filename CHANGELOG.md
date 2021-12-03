# Changelog

## 0.7.0 - 2021-12-03

### Fixed

- 修复Github CI未成功打tag分支镜像的问题

## 0.6.0 - 2021-12-03

### Added

- 新增对于Response.Header使用Go Template的支持

### Changed 

- DTO新增支持字段render_header和header_template
- Github Action升级docker/build-push-action@v2

### Fixed

- 修复创建规则时，StatusCode设置不生效的问题