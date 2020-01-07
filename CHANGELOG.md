<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [v1.0.0](#v100)
- [v1.0.0-beta.6](#v100-beta6)
- [v1.0.0-beta.5](#v100-beta5)
- [v1.0.0-beta.4](#v100-beta4)
- [v1.0.0-beta.3](#v100-beta3)
- [v1.0.0-beta.2](#v100-beta2)
- [v1.0.0-beta.1](#v100-beta1)
- [v1.0.0-beta](#v100-beta)
- [v1.0.0-alpha.4](#v100-alpha4)
- [v1.0.0-alpha.3](#v100-alpha3)
- [v0.9.8-beta](#v098-beta)
- [v0.9.8-alpha](#v098-alpha)
- [v0.9.7](#v097)
- [v0.9.7-beta](#v097-beta)
- [v0.9.7-alpha](#v097-alpha)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->


## v1.0.0

**Bug Fix:**
- Running status animation (#1351)
- Workflow required field and undefined func (#1350)
- Output-resource-path required validate (#1354)
- Overall last transition time not correct (#1360)

**Feature:**
- Cyclone web **supports show final logs of stages** (#1349)
- Cyclone web **supports show real-time logs of stages** (#1352)
- Make qps and brust configurable for k8s client (#1356)

**Docs:**
- Explain Kubernets resources used in Cyclone (#1337)

## v1.0.0-beta.6

**Bug Fix:**
- Fix svn checkout error can't convert string from 'UTF-8' to native encoding(#1341)

**Feature:**
- Feat: svn post commit trigger support selecting repo (#1343)

## v1.0.0-beta.5

**Bug Fix:**
- Deleting pod stuck on Terminating status (#1321)
- Update workflow display form is incorrect (#1309)
- Delete branch should not tirgger workflowrRun (#1336)
- Update pvc failed as old pvc watchdog have not been deleting completed (#1338)

**Feature:**
- Add WorkflowRun parallelism constraints and waiting queue (#1280)
- PR events support specify target branch (#1330)
- Record pvc usage in list executioncontexts api (#1331)

## v1.0.0-beta.4

**Bug Fix:**
- Support extension parameters for sonar scanner (#1317)
- Reduce cpu usage caused by go ticker incorrect use (#1318)

**Feature:**
- Rrecord workflow topology to workflow run (#1312)

**Others:**
- Upgrade golang to 1.12.12 & golangci-lint to 1.20.1

## v1.0.0-beta.3

**Bug Fix:**

- Resync of wfr (#1276)
- Overall status judgement when multiple stage errors (#1281)
- Close websocket when workflow run running completed (#1282)
- Eof of stage logs FolderReader (#1286)
- Workflow trigger controller should only process cron type trigger (#1307)
- Execution context of workflow trigger expire when cluster changed (#1306)

**Feature:**

- Support start and stop pvc watcher (#1290)

## v1.0.0-beta.2

**Bug Fix:**

- Overall status wrong when stage status is cancelled (#1273)

## v1.0.0-beta.1

**Features:**

- Add execution contexts API (#1266)

**Bug Fix:**

- Image tag wrong when workflow triggered by bitbucket release (#1262)
- GitHub tag trigger problem (#1269)

**Chore:**

- Enable PVC watcher (#1263)

## v1.0.0-beta

**Features:**

- Init pull request status when wfr start (#1245)

**Bug Fix:**

- Send notification only once (#1247)
- Fix git clone from bitbucket failed (#1249)
- Add target branch for list pull requests api (#1252)
- Defer exit to collect init containers' logs (#1251)

## v1.0.0-alpha.4

**Features:**

- Support dry-run for creating integration (#1199)
- Make dind bip configurable (#1200)
- Make the resource requirement of GC pod configurable (#1203)

**Bug fix:**

- Github release event webhook trigger workflow twice (#1201)
- Update gradle image version to 5.5.1 for stage template (#1204)

## v1.0.0-alpha.3

**Bug fix:**

- Fix can not get sonar scan results (#1181)
- Fix a low level bug for GitLab v3 (#1187)
- Skip ssl verify for GitLab (#1190 #1195)
- Clarify error code (#1193)
- Fix can not pull bitbucket code with with personal access token (#1197)
- Fix label value for wft event repo is too long (#1198)

**Performance:**

- Improve git resolver performance with --depth (#1177)

## v0.9.8-beta

**Bug Fix**
- Reduce cpu usage caused by go ticker incorrect use (#1322)
- Deleting pod stuck on Terminating status (#1323)

## v0.9.8-alpha

**Bug Fix**
- Close WebSocket when workflow run running completed (#1283)
- Add EOF file to stage logs folder to indicate end of stage logs(#1285)

## v0.9.7

**Bug Fix:**

- Send notification only once (#1248) 
- Git clone from bitbucket failed (#1257)
- Defer exit to collect init containers' logs (#1258)
- Add target branch for list pull requests api (#1259)
- Image tag wrong when workflow triggered by bitbucket release (#1264)
- GitHub tag trigger problem (#1270) 

**Feature:**

- Add execution contexts api (#1267)

## v0.9.7-beta

**Features:**

- **Support collect init container logs**
- Support list pull requests of scm repo

**Bug Fix:**

- Can not list working pods
- Websocket interrupted occasionally
- Workflow run status check logic
- Bitbucket pr webhook problem
- Can not list gitlab branch whose project name contains dot
- GitLab version detect problem
- Validate bitbucket token
- GitLab tag deletion should not trigger workflow

## v0.9.7-alpha

**Bug fix:**

- Skip ssl verify for gitlab(#1196)
