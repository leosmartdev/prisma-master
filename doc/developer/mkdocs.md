## Made with mkdocs

This documentation can be run locally (with live reload) or you can build a static site to host anywhere such as AWS.

To serve locally, just run (at the root of the prisma git repo):

```
make serve-docs
```

To build the static site, run:

```
make docs
```

!!! note "If you get an error on build"
    If you get an error like `[Errno 2] No such file or directory: '/Users/PilotConway/Projects/prisma/doc/linked-documents/tms/cmd/tools/tdemo/prisma'`, run `make clean` then run the `make docs` or `make serve-docs` and it should build.

### Installing mkdocs

Building and serving the docs requires mkdocs to be installed. For
Ubuntu systems, you will need to `apt-get install mkdocs` and for
macOS use `brew install mkdocs`. You will also need to get the material
theme we are using. `pip install mkdocs-material` or `pip3 install mkdocs-material`.

There are a number of markdown extensions mkdocs uses for things like code
coloring, todo's, and notes that you will need to install as well. `pip3 install pymdown-extensions`.

For pdf generation of docs, we also need to install cairo and pango so the `mkdocs-pdf-export-plugin`
extension can function.

```bash tab="macOS"
brew install mkdocs cairo pango
pip install mkdocs-material pymdown-extensions mkdocs-pdf-export-plugin
```

```bash tab="Ubuntu"
sudo apt install shared-mime-info
pip3 install mkdocs mkdocs-material pymdown-extensionsn
```

## Project layout

    mkdocs.yml    # The configuration file for this documentation
    docs/
        index.md  # The documentation homepage.
        linked_documents/  # Folder containing symbolic links to documents outside doc in the repo. eg READMEs
        ...       # Other markdown pages, images and other files

## Writing documentation

All documents are in standard markdown. We also enable some plugins (as seen in the mkdocs.yml markdown_extensions settings).
These plugins enable special features like the notes, todo, warning sections, and better code highlighting.

* [All available extensions](https://python-markdown.github.io/extensions/)
* [Pymdown extensions](https://facelessuser.github.io/pymdown-extensions/)
* [Formatting notes, todo, warning, etc...](https://python-markdown.github.io/extensions/admonition/)
* [Multi-language code blocks](https://facelessuser.github.io/pymdown-extensions/extensions/superfences/)
* [Task lists](https://facelessuser.github.io/pymdown-extensions/extensions/tasklist/)

### Notes

```
!!! note
    This is a note
```

!!! note
    This is a note


```
!!! note "With Custom Title"
    This is a note
```

!!! note "With Custom Title"
    This is a note

### Todo

```
!!! todo
    This is a TODO
```

!!! todo
    This is a TODO

### Example

```
!!! example
    ```
    EXAMPLE CODE
    ```

    ??? success
        ```
        Output on success
        ```

    ??? failure
        ```
        Output on failure
        ```
```

!!! example
    ```
    EXAMPLE CODE
    ```

    ??? success
        ```
        Output on success
        ```

    ??? failure
        ```
        Output on failure
        ```


For full documentation visit [mkdocs.org](https://mkdocs.org).
