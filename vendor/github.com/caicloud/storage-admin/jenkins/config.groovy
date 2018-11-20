env.PRIVATE_CARGO_REGISTRY = "cargo.caicloudprivatetest.com"
env.IMAGE_NAME_ADMIN = "caicloud/storage-admin"
env.IMAGE_NAME_ADMIN_CONTROLLER = "caicloud/storage-admin-controller"
env.IMAGE_NAME_ADMISSION = "caicloud/storage-admission"
env.IMAGE_NAME_CONTROLLER = "caicloud/storage-controller"
env.DOCKERFILE_ADMIN = "./build/admin/Dockerfile"
env.DOCKERFILE_ADMIN_CONTROLLER = "./build/admin-controller/Dockerfile"
env.DOCKERFILE_ADMISSION = "./build/admission/Dockerfile"
env.DOCKERFILE_CONTROLLER = "./build/controller/Dockerfile"

VERSION = "v0.1.9"
PUBLISH = false

if ("${env.BRANCH_NAME}".toLowerCase().contains("pr-")) {
    env.IMAGE_TAG = "${env.BRANCH_NAME}-${env.BUILD_NUMBER}".toLowerCase()
} else if ("${env.BRANCH_NAME}" == "master") {
    if (PUBLISH) {
        env.IMAGE_TAG = "${VERSION}"
    } else {
        env.IMAGE_TAG = "${VERSION}-snapshot-${env.BUILD_NUMBER}"
    }
} else {
    env.IMAGE_TAG = "${env.BRANCH_NAME}-rc${env.BUILD_NUMBER}"
}
