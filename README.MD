# Introduction

This repository is an example of how to implement a webhook for Port for creation and deletion of developer resources (i.e. s3 buckets, sns topics, etc).  
It uses a simple web server that gets webhook requests from Port (see [docs](https://docs.getport.io/platform-overview/self-service-actions/setting-self-service-actions-in-port)) and applies user inputs to Terraform files. You can fork this repository to extend its core functionality or add more templates as your use-case requires.

## Architecture Diagram

![architecture](diagram.drawio.png)

## Example setup (S3 bucket)

The following will walk you through how to set up Port actions to create and delete S3 buckets (a similar flow is required for resources on other clouds).
>**_NOTE:_**  
>For this tutorial you'll need:
>1. an AWS account with permissions to create S3 buckets
>1. an account in [PORT](https://getport.io)

First, we want to run this web server and make it public for Port.  
To make your local machine public to Port you can use either [ngrok](https://ngrok.com/) or [smee](https://smee.io/).  
In the following example I'll use ngrok, run:
```shell
ngrok http 8080
```
and you should get a forwarding URL:
```shell
ngrok

Session Status                online                                                                                       
Account                       <Account Name>                                                                                                        
Version                       3.1.0                                                                                             
Region                        Europe (eu)                                                                    
Latency                       68ms                                                                                                                           
Web Interface                 http://127.0.0.1:4040                                                                       
Forwarding                    https://1234-5678-9101-112-1314-1516-abcd-efgh-ijkl.eu.ngrok.io -> http://localhost:8080           
                                                                                                                
Connections                   ttl     opn     rt1     rt5     p50     p90                                                                
                              79      0       0.00    0.00    4.91    31.99                                               
```
Keep the forwarding URL and we will use it later.  
  
Next, we will deploy this project as a docker container, run the following:
```shell
docker run \
  -e PORT_CLIENT_ID=<PORT_CLIENT_ID> \            
  -e PORT_CLIENT_SECRET=<PORT_CLIENT_SECRET> \
  -e AWS_ACCESS_KEY_ID=<AWS_ACCESS_KEY_ID> \
  -e AWS_SECRET_ACCESS_KEY=<AWS_SECRET_ACCESS_KEY> \
  -e AWS_SESSION_TOKEN=<AWS_SESSION_TOKEN> \
  -e DEBUG=true \
  -p 8080:8080 \
  -it hedwigz/webhook-terraform:v0.1
```
Now we have a running server that gets webhook request from Port and applies them to terraform files!  
  
Next, let's set up a blueprint for an S3 bucket on Port. This can be achieved by using the web ui or using the [Port's terraform provider](https://docs.getport.io/integrations/terraform). We will stick to the web ui for now - go to Blueprints page and create a new blueprint

![create new blueprint](images/newBlueprintButton-3229130bc81c1ded3d6d9f929aa40897.png)

You can create how many properties you wish to see on each bucket, but for this example we will keep it lean with two properties - `URL` and `Bucket Name`.
Enter the following JSON:
```json
{
    "identifier": "s3_bucket",
    "title": "AWS Bucket",
    "icon": "Bucket",
    "schema": {
        "properties": {
            "bucket": {
                "type": "string",
                "title": "Bucket Name"
            },
            "url": {
                "type": "string",
                "title": "URL",
                "format": "url"
            }
        },
        "required": [
            "url",
            "bucket"
        ]
    },
    "mirrorProperties": {},
    "formulaProperties": {},
    "relations": {}
}
```
Press the `save` button and you should see the Bucket blueprint added to the screen:
![blueprint created](images/Screen%20Shot%202022-09-19%20at%2019.25.03.png)
Next, we want to create an action that will create buckets from the UI. For this, expand the AWS Bucket blueprint and press the `Edit actions` button:
![Edit actions](images/Screen%20Shot%202022-09-19%20at%2019.28.47.png)
Then, paste the following action configuration:
> **_NOTE:_**  replace `<your ngrok forwarding URL>` with the URL you got earlier
```json
[
    {
        "identifier": "create_bucket",
        "title": "Create Bucket",
        "icon": "Bucket",
        "userInputs": {
            "properties": {
                "bucket_name": {
                    "type": "string",
                    "title": "Name for the S3 bucket"
                },
                "tags": {
                    "type": "object",
                    "title": "Tags",
                    "default": {}
                }
            },
            "required": [
                "bucket_name"
            ]
        },
        "invocationMethod": {
            "type": "WEBHOOK",
            "url": "<your ngrok forwarding URL>"
        },
        "trigger": "CREATE",
        "description": "Create a new S3 Bucket in AWS"
    }
]
```

It's time to test our action, go to `AWS Buckets` page on Port and trigger the `CREATE` action:
![Create Bucket](images/Screen%20Shot%202022-09-20%20at%2014.05.53.png)

Fill in the form and press `Create`:
![Create Bucket Form](images/Screen%20Shot%202022-09-20%20at%2014.07.34.png)