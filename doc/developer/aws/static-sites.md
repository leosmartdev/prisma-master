# Static Sites

## Sites

We currently host three domains as static sites in AWS. A site for all our documentation, storybook
sites, and a site that hosts our release binaries. All sites are hosted using only `https`
through CloudFront.

### documentation.mcmurdo.io

This site hosts our documentation site, including the entire `doc` directory in orolia/prisma,
the release site from the orolia/prisma-release-documentation repo, and in the future, the docs 
directory from the orolia/prisma-ui repo. We
also include other documentation like the storybook builds for `prisma-electron`.

The documentation site is hosted in the S3 bucket `prisma-documentation`. The root of that directory
contains an index.html file that when loaded redirects the browser to `latest/index.html`. 

!!! note TODO
    In the future, we shouldn't be redirecting like this, instead, we should have a lambda function 
    in clountfront that properly redirects in AWS to the latest documentation folder. 
    
Also in the root of that folder are the following: 

  * godoc
  * releases
  * nightly
  * latest
  * A folder for each version released. 

#### latest

This is the build of [orolia/prisma `/doc` folder on the master branch](https://github.com/orolia/prisma/tree/master/doc)
on the master branch. On every push to master, CodeBuild will build the docs site and then sync it 
to the `latest` folder. 

This is the default site that is loaded when a user goes to 
[https://documentation.mcmurdo.io](https://documentation.mcmurdo.io). 

#### nightly 

This is the build of [orolia/prisma `/doc` folder on the develop branch](https://github.com/orolia/prisma/tree/develop/doc). 
On every push to develop, CodeBuild will build the docs site and then sync it to the `nightly` 
folder.

You can get to this documentation at the url [https://documentation.mcmurdo.io/nightly](https://documentation.mcmurdo.io/nightly)

#### Version folders 

During the release build of [orolia/prisma](https://github.com/orolia/prisma) when a new version is 
tagged, the CodeBuild project will build and deploy the `/doc` folder and sync it to a folder that 
matches the tag name. So for every version you can access the documentation site. 

You can get to this documentation at a url like `https://documentation.mcmurdo.io/<VERSION>`
or example [https://documentation.mcmurdo.io/1.7.2](https://documentation.mcmurdo.io/1.7.2)

#### godoc

We also publish the `pkg` directory of godoc during the builds. This allows developers to see a nice 
set of documentation for all the go code.

The godoc folder follows the same convention as the top level, so under godoc are the following 
copies of the generated documentation: 

  * `godoc/latest`: build on every push to master
  * `godoc/nightly`: build on ever push to develop
  * `godoc/<VERSION>`: build of every tag

#### releases

`releases` is a copy of the release site, which is build from the repo
`prisma-release-documentation`. This is a complete static site just like the top level but contains
release information and links to the release binaries hosted on `releases.mcmurdo.io`. This is in
a separate repo other than `prisma` because we need to be able to edit the releases separate from
the code as the release pages rely on the built binaries.

On every push to `master` in [https://github.com/orolia/prisma-release-documentation](https://github.com/orolia/prisma-release-documentation)
the release site is deployed here by CodeBuild. 

### releases.mcmurdo.io

This is a domain that is directed at the S3 bucket `prisma-c2/releases`. It hosts all the binary
builds for the releases so we can easily download them using something like curl and basic http auth
strings.

### storybook.mcmurdo.io

Hosts the storybook stories from the prisma-ui repository. This site is hosted from the bucket 
`s3://prisma-storybook`. At the top level are folders for each package that has a storybook.

We currently host storybooks for the following packages: 

  * [https://storybook.mcmurdo.io/prisma-electron/latest](https://storybook.mcmurdo.io/prisma-electron/latest)
  * [https://storybook.mcmurdo.io/prisma-map/latest](https://storybook.mcmurdo.io/prisma-map/latest)
  * [https://storybook.mcmurdo.io/prisma-ui/latest](https://storybook.mcmurdo.io/prisma-ui/latest)
  * [https://storybook.mcmurdo.io/prisma-form](https://storybook.mcmurdo.io/prisma-form)
  
For prisma-electron, prisma-map, and prisma-ui the folders for each contain the same `latest`, 
`nightly`, and `<VERSION>`. For prisma-form just the `master` branch is deployed currently. 

## Creation and Setup

We host our static sites by exposing an S3 bucket configured as a static host to CloudFront
where we serve the files and enforce a BASIC HTTP authentication password for the site.

### Creating a Static Site

To create a static site, log into the console in S3 and select the bucket (or create) that you want to host the site. Then select that bucket. At the top should be tabs, select the Properties tab.

There should be a grid of squares, find the `Static website hosting` and click on it. Select `Use this bucket to host a website` and be sure to set the index document and error document. Copy the URL thats provided at the top of this pane. Click save.

Now, if you navgivate to the endpoint that was shown in the that square, you should see that S3 is hosting your website, serving that index document you selected.

At this point for S3 to serve the website you will have to make the bucket public. You can avoid this by using CloudFront and subsequent sections below to lock down who can access this site.

!!! note

    The index document is just the general filename, so it should usually be `index.html`. This is not the exact filename, it just tells AWS to look for a file named `index.html` whenever a specific file name is not requested. `http://bucket.s3-website.amazonaws.com/releases/foo/` would resolve the file `releases/foo/index.html`.

### Setting up CloudFront to serve the site

Now that S3 is serving the site, we can configure CloudFront to serve the bucket in its system.

We use CloudFront for two reasons. First, it allows us to use a custom domain (not just forwarding) for the site. Second, it allows us to add protections to only allow authorized access to the site.

Go to the CloudFront site in AWS. Click on Create Distribution (Blue button at the top). Then, in the `Web` section click `Get Started`.

For origin domain name, paste the URL you copied above.

!!! warning
    Do not use the dropdown to the select the bucket, be sure to paste the URL to the S3 public site. If you select S3 bucket, CloudFront will not resolve the `index.html` and much of your site will be broken. By using the S3 hosted site URL, S3 will resolve when a non .html path is selected and do what you are expecting.

Set the HTTPS settings as you wish (usually Redirect HTTP to HTTPS).

Under `Distribution Settings`, there is a text field for `Alernate Domain Names`. Here, put the subdomain of `mcmurdo.io` you want to use, for example, `releases.mcmurdo.io`. If you don't do this, then the subdomain will not work as CloudFront will reject requests using that domain.

Select Custom SSL certificate and at the dropdown select `mcmurdo.io` certificate. (NOT the test certificate).

If you want, you can enable logging to an S3 bucket.

Once complete, click `Create Distribution` at the bottom.

At this point, it will take a bit, but CloudFront will now be serving your site at the URL provided by CloudFront. You can navigate your browser there and see the same site as S3 was hosting in the previous section.

### Setting up the Subdomain

Now that CloudFront is hosting your site, you can setup the subdomain in Route 53 to point to the cloud front site.

Navigate to Route53 in AWS and go to the Hosted Zones and select `mcmurdo.io`. At the top click `Create Record Set`. Type in the subdomain name you wish to use and make sure `A` record is selected. Now under Alias hit `Yes`.

Provided you filled in the `Alternate Domain Names` section correctly when setting up CloudFront, then the cloud front distrubtion should show up in the dropdown. Select it then hit `Create` at the bottom of the page.

After some time propogating, your site should now be available on that domain.

### Locking down the S3 Bucket

Now that your static site is up and running, we need to lock down the S3 bucket so only CloudFront can access the files. This way, if someone knows the S3 bucket they can't bypass the CloudFront site to get to it (especially important if you are password protecting the site using CloudFront in the section below).

This is done by changing the Bucket policy in S3 so that the files are publicly accessibly for read ONLY when a specific token is passed as a header. The token will be hard coded in the bucket policy, but only CloudFront and S3 policy will know the token. Any other request (eg from a browser) directly to S3 will result in permission denied.

The following link describes how this works and that it's the current official way of doing this by AWS.

https://serverfault.com/questions/791658/can-i-hide-s3-and-cloudfront-endpoints

!!! note
    If needed, you can change the Token in the policy and CloudFront periodically to ensure security, but it will need to be done manually.

To set this up, go to S3, and select your bucket then click on `Permissions` tab at the top. There will be a button called `BucketPolicy`, click that.

Now, you may have some policies already in the editor shown. The key here is none of them should be Public for s3:GetObject except the one we are creating, otherwise, the other policy may take precendence. Specific policies listed here are ok for ARNs or other users.

Paste the following policy in to the `Statement` array replacing `<BUCKETNAME>` with your bucket name (eg prisma-documentation) and `<TOKEN>` with a long string that is the security token. You can use a password generator if you wish to generate a long random string 30 characters minimum please.

```json
{
  "Sid": "PublicReadForGetBucketObjects",
  "Effect": "Allow",
  "Principal": "*",
  "Action": "s3:GetObject",
  "Resource": "arn:aws:s3:::<BUCKETNAME>/*",
  "Condition": {
      "StringEquals": {
          "aws:Referer": "<TOKEN>"
      }
  }
}
```

Now, head to CloudFront in AWS and select the CloudFront distribution, then click `Origins and Origin Groups` tab. Click on the Origin (there should only be one) then click Edit.

At the bottom should be `Origin Custom Headers`. For the name type in `Referred` and then as the value past the key you set in the Policy above.

### Password protecting CloudFront

To password protect CloudFront, we need to use a Lambda function to process the Basic HTTP auth (or if you want, some other auth mechanism, but that hasn't been tested by us yet).

We have an existing Auth lambda function, so if you are ok using the same password as `documentation.mcmurdo.io` and `releases.mcmurdo.io` then you can just use that Lambda function directly. Otherwise, copy the existing lambda function and change the username.password hard coded in that function to the values you wish to use.

The lambda function is located here: [staticSiteBasicHttpAuthentication](https://console.aws.amazon.com/lambda/home?region=us-east-1#/functions/staticSiteBasicHttpAuthentication?tab=graph)

To connect the lambda function to your CloudFront distribution, naviagate to the `Behaviors` tab in CloudFront, then select edit. At the bottom is a section for `Lambda function Associations`.

The CloudFront Event is `Viewer Request` then past the ARN for your lambda function in the next box. Select `include Body` then hit `Yes, Edit` at the bottom.

Now for every request to view, the lambda function will verify the user has credentials to view the page before serving.
