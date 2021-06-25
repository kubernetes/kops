# Updating The Default Base AMI

With the release of `kOps 1.18`, the base AMI was switched over from a pre-baked `kOps` AMI to the official Ubuntu 20.04 LTS image.
This makes the image update process easier, because kOps contributors no longer have to fully build and test a full AMI, but rather base off of the latest stable Ubuntu image.
In order to make sure we're up to date with the latest releases, we regularly follow the official [Ubuntu EC2 AMI Locator](https://cloud-images.ubuntu.com/locator/ec2/) website and update to the latest version of `focal` which is available across **all AWS regions**, including `gov` and `cn`, for full support.

The process of updating the AMI version is as following:

- Find the most recent release on the official [Ubuntu EC2 AMI Locator](https://cloud-images.ubuntu.com/locator/ec2/). Make sure it's available across all regions. The ones with the slowest release cycle are usually `gov`, `ap` and `cn` ones, so the best option would usually be to take the most recent release from one of these regions.
- Replace the timestamp on this line in the `alpha` channel, where the `ubuntu` image is referred. [Example](https://github.com/kubernetes/kops/blob/25eb1c98225450bed82d38e52d150d7a69a2c95a/channels/alpha#L47).

    !!!note
        Before updating `alpha` channel, check and see if `alpha` and `stable` channel are both running the same AMI version. If `stable` currently runs a different version, and more than 7-10 days passed since `alpha` was updated- it's safe to also push the version currently in `alpha`, to `stable` in the same PR.
        e.g., let's say that the most recent available on Ubuntu image locator is `20201210`, `alpha` is currently using `20201101` and `stable` is currently using `20201015`. If `alpha` was updated at least 7-10 days prior to your desired change, you can update `stable` with the version that was listed in `alpha` **before** your change. Then- you may update `alpha` with the most recent version of Ubuntu.

    !!!note
        When updating the `stable` channel with a new ami version, there's a pretty good chance that this will cause some tests to fail. Thus, it's worth running `hack/update-expected.sh`. This will update all the integration tests with the newly updated ami. To get ahead of this locally before pushing - `make test` will confirm that everything is updated as should.

- Let the new AMI version bake-in in `alpha` channel for at least 7-10 days, afterwhich it's safe to create a follow-up PR to push the latest version to `stable` channel.  
