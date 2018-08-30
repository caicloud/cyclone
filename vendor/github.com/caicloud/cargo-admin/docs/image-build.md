# Image Build

Image build is another way to upload docker images. It builds images from Dockerfile and necessary context, and then push images to registry. 

We have provide Dockerfile templates for different applications. For the moment, they are:

- Tomcat
- Python
- NodeJS
- PHP

Suppose you are working on a Python 2.7 project, you can choose a Dockerfile template based on `python:2.7-alpine` base image.

```Dockerfile
FROM python:2.7-alpine  
  
# Create working directory  
WORKDIR /usr/src/app  
  
# Add your uploaded file to workplace, if it's an archive file (zip, tar, tar.gz),  
# it will be unpacked and added to worksapce.  
ADD <uploaded-file> .  
  
# Make sure the requirements.txt file exists and then install dependencies using it.  
RUN touch requirements.txt  
RUN pip install --no-cache-dir -r requirements.txt  
  
CMD [ "python", "./<entrypoint>.py" ]
```

Then what you need to do is populate the `<uploaded-file>` with the file name you uploaded and `<entrypoint>` with your entry point Python script. By all this, your docker images will be built and pushed to docker registry.

## Uploaded Archive File Structure

Uploaded file supports both single file and archive file (`zip`, `tar`, `tar.gz`). For example, if you are using tomcat to deploy a web application, upload your `war` file directly. And if your are working on a NodeJS project, you need to pack all your files before upload.

How your files are organized in archive file is important. Otherwise you need more changes of the Dockerfile. It's strongly recommended that you _pack your project files without nesting directories_. For example, suppose your have following project files:

```bash
$ tree
.
├── README.md
├── app
├── docs
├── requirements.txt
├── setup.py
```

You should pack them like:

```bash
$ tar czvf app.tar.gz ./*
```

## How To Access Uploaded Files In Dockerfile

Put in simply, you can regard `WORKDIR` in Dockerfile as same to your project root directory. So if you have following files, and pack them correctly, then you can access these files like: `./setup.py`, `setup.py` in Dockerfile.

```bash
$ tree
.
├── README.md
├── app
├── docs
├── requirements.txt
├── setup.py
```

And also, please note that, your Dockerfile would go into the same directory, so the final file structure docker daemon seems is much like:

```bash
$ tree
.
├── Dockerfile
├── README.md
├── app
├── docs
├── requirements.txt
├── setup.py
```