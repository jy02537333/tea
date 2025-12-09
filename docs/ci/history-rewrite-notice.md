# 历史重写维护窗口通知模板 / Maintenance Window Notice Templates

本文件提供在对 `master` 分支进行历史重写前、中、后的对外通知模版（中/英）。可按需替换尖括号占位符。

---

## 维护前通知（中文）

> 标题：关于 `master` 分支历史清理维护的通知（[日期 时间]）
>
> 各位同事好：
>
> 为移除仓库历史中的超大文件与敏感信息，并解除 GitHub 推送限制，我们计划在 [日期 时间] 对 `master` 分支进行一次性历史重写。维护窗口预计 [X 分钟]，期间请避免向 `master` 推送或合并变更。
>
> 影响与说明：
>
> - 维护将修改 `master` 的提交历史，完成后需重新克隆或执行硬重置对齐新历史；
> - 维护期间将临时允许强制推送（--force-with-lease）；
> - 不涉及业务逻辑修改，不影响构建与运行。
>
> 协作者对齐指引（维护完成后执行）：
>
> ```bash
> git fetch origin
> git checkout master
> git reset --hard origin/master
> ```
>
>
> 如有疑问或特殊安排，请在维护前联系 <负责人/联系人>。感谢理解与配合！

## 维护开始通知（中文）

> 标题：开始进行 `master` 历史清理维护（预计 [X 分钟]）
>
> 历史清理维护已经开始，请在维护完成前避免向 `master` 推送或合并。
>
> 维护完成后将第一时间通知，并附上对齐新历史的步骤。

## 维护完成通知（中文）

> 标题：`master` 历史清理维护已完成（请对齐新历史）
>
> 历史清理已完成。请按以下命令对齐到新历史：
>
> ```bash
> git fetch origin
> git checkout master
> git reset --hard origin/master
> ```
>
> 若需继续未完成分支开发，可先导出补丁再应用：
>
> ```bash
> git format-patch origin/master..my/feature -o /tmp/patches
> git am /tmp/patches/*.patch
> ```
>
> 如遇任何问题，请联系 <负责人/联系人>。

---

## Pre-maintenance Notice (English)

> Subject: Scheduled maintenance for `master` history rewrite ([Date Time])
>
> Hello team,
>
> To remove oversized blobs and sensitive content from the repository history and to clear GitHub push restrictions, we will perform a one-time history rewrite on `master` at [Date Time]. The maintenance window is expected to last [X minutes]. During this time, please avoid pushing or merging into `master`.
>
> Impact & Notes:
>
> - This will rewrite the commit history of `master`. After completion, please re-clone or hard reset to align with the new history.
> - A force-with-lease push will be performed during the window.
> - No application logic changes; builds and runtime should remain unaffected.
>
> Alignment steps (after maintenance):
>
> ```bash
> git fetch origin
> git checkout master
> git reset --hard origin/master
> ```
>
>
> For any questions or special arrangements, please contact <Owner/POC> prior to the window. Thank you!

## Maintenance Start (English)

> Subject: `master` history rewrite started (ETA [X minutes])
>
> The maintenance has started. Please avoid pushing/merging to `master` until completion.
>
> We will notify once it is complete along with alignment steps.

## Maintenance Complete (English)

> Subject: `master` history rewrite completed (please align your local copy)
>
> The maintenance has completed. Please align your local copy with the following:
>
> ```bash
> git fetch origin
> git checkout master
> git reset --hard origin/master
> ```
>
> If you need to continue work on a pre-existing branch, you can export and apply patches:
>
> ```bash
> git format-patch origin/master..my/feature -o /tmp/patches
> git am /tmp/patches/*.patch
> ```
>
> Please reach out to <Owner/POC> if you encounter any issues.
