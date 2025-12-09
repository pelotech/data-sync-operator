# Changelog

## [0.1.3](https://github.com/pelotech/data-sync-operator/compare/0.1.2...0.1.3) (2025-12-09)


### Bug Fixes

* enable event rbac ([#59](https://github.com/pelotech/data-sync-operator/issues/59)) ([b854f70](https://github.com/pelotech/data-sync-operator/commit/b854f7016903d8e99eff05b2821accd8609b6d8b))

## [0.1.2](https://github.com/pelotech/data-sync-operator/compare/0.1.1...0.1.2) (2025-12-04)


### Bug Fixes

* release please recognize chart changes ([#55](https://github.com/pelotech/data-sync-operator/issues/55)) ([80588a3](https://github.com/pelotech/data-sync-operator/commit/80588a3d7d1aebc663bc098fd8c56239fa61d2cc))
* set keep in chart to false as it is more sensible default ([#54](https://github.com/pelotech/data-sync-operator/issues/54)) ([8a27a1e](https://github.com/pelotech/data-sync-operator/commit/8a27a1e4fe2d8ee433e5e080e45f840e55895300))

## [0.1.1](https://github.com/pelotech/data-sync-operator/compare/0.1.0...0.1.1) (2025-12-04)


### Bug Fixes

* give vmdi controller delete collection permission ([#52](https://github.com/pelotech/data-sync-operator/issues/52)) ([0804f83](https://github.com/pelotech/data-sync-operator/commit/0804f838c86bc487bfdae93760b265d0c225c14b))

## [0.1.0](https://github.com/pelotech/data-sync-operator/compare/0.0.1...0.1.0) (2025-12-02)


### Features

* vm disk image mvp implementation ([#10](https://github.com/pelotech/data-sync-operator/issues/10)) ([1b02b1a](https://github.com/pelotech/data-sync-operator/commit/1b02b1abea3175717ec0951096df9cbc69c472aa))


### Bug Fixes

* **deps:** update golang docker tag to v1.25 ([62e7e7d](https://github.com/pelotech/data-sync-operator/commit/62e7e7de8a15adde271a5e5f95bf4c6d99e1168f))
* **deps:** update k8s.io/utils digest to bc988d5 ([#14](https://github.com/pelotech/data-sync-operator/issues/14)) ([0060f2f](https://github.com/pelotech/data-sync-operator/commit/0060f2fbfd9c02420144933138e679fa645241a4))
* **deps:** update kubernetes packages to v0.34.2 ([#1](https://github.com/pelotech/data-sync-operator/issues/1)) ([b61d5ec](https://github.com/pelotech/data-sync-operator/commit/b61d5ecc5226fd4623d3f67cb9836219902df928))
* **deps:** update module github.com/kubernetes-csi/external-snapshotter/client/v6 to v8 ([#16](https://github.com/pelotech/data-sync-operator/issues/16)) ([f9cc1f4](https://github.com/pelotech/data-sync-operator/commit/f9cc1f4146f0370eaaad85bcb1e8fd904a2262f7))
* **deps:** update module github.com/kubernetes-csi/external-snapshotter/client/v6 to v8 ([#22](https://github.com/pelotech/data-sync-operator/issues/22)) ([f924f4a](https://github.com/pelotech/data-sync-operator/commit/f924f4acc136e738caa9d705ffc464949f859304))
* **deps:** update module go.uber.org/zap to v1.27.1 ([#15](https://github.com/pelotech/data-sync-operator/issues/15)) ([9a593a8](https://github.com/pelotech/data-sync-operator/commit/9a593a8d9cb442f4fedaeac5419a3978cffe6afd))
* **deps:** update module sigs.k8s.io/controller-runtime to v0.22.4 ([#6](https://github.com/pelotech/data-sync-operator/issues/6)) ([6bf3daa](https://github.com/pelotech/data-sync-operator/commit/6bf3daa3e4b91b7291e97e0fb94f95d2ac1bd67f))
* remove default to dev mode config ([#18](https://github.com/pelotech/data-sync-operator/issues/18)) ([ffa3a7a](https://github.com/pelotech/data-sync-operator/commit/ffa3a7a40e1af09277e83da98f8f2cd5d8aae29e))
* rename the chart so it doesn't conflict with the repo name ([#32](https://github.com/pelotech/data-sync-operator/issues/32)) ([335116b](https://github.com/pelotech/data-sync-operator/commit/335116b0091549091c76eae95302bc1b5e629a8a))
* **renovate:** don't always use fix for deps updates ([36ba49f](https://github.com/pelotech/data-sync-operator/commit/36ba49fff0d118b032f44aa80fdca0902ca0128a))


### Refactors

* restructure charts structure and release-please artifacts publishing ([#47](https://github.com/pelotech/data-sync-operator/issues/47)) ([caac69d](https://github.com/pelotech/data-sync-operator/commit/caac69d8e4eba6ed2a74ca310de29de41ffbed6c))


### Chores

* add air to hot reload the operator for local development ([#11](https://github.com/pelotech/data-sync-operator/issues/11)) ([74a3238](https://github.com/pelotech/data-sync-operator/commit/74a32387eebe4f5ddac07c3f1a5bdfa3d97054c1))
* configure release please to do chart releases ([#24](https://github.com/pelotech/data-sync-operator/issues/24)) ([8c1f881](https://github.com/pelotech/data-sync-operator/commit/8c1f8819760af02d67c808e57e00635b885569cf))
* fix chart name issue ([#33](https://github.com/pelotech/data-sync-operator/issues/33)) ([88bcfd3](https://github.com/pelotech/data-sync-operator/commit/88bcfd3ea1090c491d8f1d6e0c9d567f1054f6f5))
* fix var mismatch, reverted package strings in workflow checks, fix readme generation loop ([#36](https://github.com/pelotech/data-sync-operator/issues/36)) ([4cf71cc](https://github.com/pelotech/data-sync-operator/commit/4cf71cc86d6223ba6a4038dd0d520f8ce609db1b))
* helm chart ([#17](https://github.com/pelotech/data-sync-operator/issues/17)) ([b81ad2c](https://github.com/pelotech/data-sync-operator/commit/b81ad2cedd754e1be9844022e803edcc5792a981))
* **main:** release data-sync-operator 1.0.0 ([#25](https://github.com/pelotech/data-sync-operator/issues/25)) ([b3e498f](https://github.com/pelotech/data-sync-operator/commit/b3e498fe901e8a8196c2e1cf1243ac230bb28728))
* **main:** release data-sync-operator-chart 1.0.1 ([#28](https://github.com/pelotech/data-sync-operator/issues/28)) ([3cb4431](https://github.com/pelotech/data-sync-operator/commit/3cb4431ec2d7d27d376b31f1389d92558be1d7a1))
* **main:** release data-sync-operator-chart 1.0.1 ([#46](https://github.com/pelotech/data-sync-operator/issues/46)) ([1328cdc](https://github.com/pelotech/data-sync-operator/commit/1328cdce9ae5e6d13f9550390952735b7b3a1abd))
* **main:** release data-sync-operator-chart 1.0.2 ([#31](https://github.com/pelotech/data-sync-operator/issues/31)) ([9078e2d](https://github.com/pelotech/data-sync-operator/commit/9078e2da4a0bb13182a5ab8fcbd48203185aa902))
* **main:** release data-sync-operator-chart 1.0.3 ([#34](https://github.com/pelotech/data-sync-operator/issues/34)) ([7c6c118](https://github.com/pelotech/data-sync-operator/commit/7c6c118f771ee8ce7c799b721377dc9136c0c72b))
* **main:** release data-sync-operator-chart 1.0.4 ([#35](https://github.com/pelotech/data-sync-operator/issues/35)) ([65ab562](https://github.com/pelotech/data-sync-operator/commit/65ab5623b50709051e96d1997e52e2c7d5325f7a))
* **main:** release data-sync-operator-chart 1.0.5 ([#37](https://github.com/pelotech/data-sync-operator/issues/37)) ([10e60a2](https://github.com/pelotech/data-sync-operator/commit/10e60a293237c98c2c8b72583cbeeda85cb4a502))
* **main:** release data-sync-operator-chart 1.0.6 ([#40](https://github.com/pelotech/data-sync-operator/issues/40)) ([00ae033](https://github.com/pelotech/data-sync-operator/commit/00ae033d4349be39dd38de7d232fa209a27c7463))
* **main:** release data-sync-operator-chart 1.0.7 ([#42](https://github.com/pelotech/data-sync-operator/issues/42)) ([b1fb083](https://github.com/pelotech/data-sync-operator/commit/b1fb0835d49081bde714e822d6a27c764ad765d6))
* **main:** release data-sync-operator-chart 1.0.8 ([#44](https://github.com/pelotech/data-sync-operator/issues/44)) ([56d5966](https://github.com/pelotech/data-sync-operator/commit/56d59666d548528fb15ec24d339cdfd5b3382318))
* manually bump chart readme version ([78c0e80](https://github.com/pelotech/data-sync-operator/commit/78c0e80b7225269f1165f96b4d9ffdab265d1df8))
* readme and action tuning ([#43](https://github.com/pelotech/data-sync-operator/issues/43)) ([8e6dc16](https://github.com/pelotech/data-sync-operator/commit/8e6dc16aa13adacdca7009b834cda088e4489b1b))
* release please test ([#39](https://github.com/pelotech/data-sync-operator/issues/39)) ([48aeac5](https://github.com/pelotech/data-sync-operator/commit/48aeac5daae5aaecc53afe7f44d88042987f9d8c))
* revert release please ([#45](https://github.com/pelotech/data-sync-operator/issues/45)) ([021b7ea](https://github.com/pelotech/data-sync-operator/commit/021b7ea2c813be67642e2fb55b64c2460e22ed87))
* update actions to handle chart and app releases ([#26](https://github.com/pelotech/data-sync-operator/issues/26)) ([281e9a4](https://github.com/pelotech/data-sync-operator/commit/281e9a4712f384b999e5fdf3b28ff3fea619d701))
* update chart repository ([#41](https://github.com/pelotech/data-sync-operator/issues/41)) ([b2982ca](https://github.com/pelotech/data-sync-operator/commit/b2982ca38288311b3a52f4e0a80983fd0545eab0))
* use path not chart_path ([#30](https://github.com/pelotech/data-sync-operator/issues/30)) ([1d1f4a3](https://github.com/pelotech/data-sync-operator/commit/1d1f4a392da0d59ace3e1bbb4d0bc18827048bb4))


### Docs

* readme clean up ([#21](https://github.com/pelotech/data-sync-operator/issues/21)) ([6c36e16](https://github.com/pelotech/data-sync-operator/commit/6c36e16f8a0da051c5aef324d6ac9e4756792a0c))
* spelling ([c575c9d](https://github.com/pelotech/data-sync-operator/commit/c575c9dc32409a076c07c8738fb1b7c6584baaac))
* update proposal ([c312180](https://github.com/pelotech/data-sync-operator/commit/c312180b6fe1c7c1e60ac6d74622f5c76f6909e4))
