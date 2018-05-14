# Creating a release

These releases represent an upstream release and not a release of the product.

- Ensure you are on the master branch
- Ensure git is not in a dirty state (ie stuff needs to be committed )
- Create the next logical tag (v0.0.3) ```git tag ``` will show you the current tags. ```git tag v0.0.3``` will create the tag.
- Create and export a valid [github token](https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/) ```export GITHUB_TOKEN=<token>```
- Run ```make release```

The end result of this should be a new release on the github page https://github.com/aerogear/mobile-cli/releases with all of the 
commits since the previous tag.
