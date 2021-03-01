load("@bazel_gazelle//:deps.bzl", "go_repository")

def go_repositories():
    go_repository(
        name = "co_honnef_go_tools",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "honnef.co/go/tools",
        sum = "h1:3JgtbtFHMiCmsznwGVTUWbgGov+pVqnlf1dEJTNAXeM=",
        version = "v0.0.1-2019.2.3",
    )

    go_repository(
        name = "com_github_agnivade_levenshtein",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/agnivade/levenshtein",
        sum = "h1:3oJU7J3FGFmyhn8KHjmVaZCN5hxTr7GxgRue+sxIXdQ=",
        version = "v1.0.1",
    )

    go_repository(
        name = "com_github_airbrake_gobrake",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/airbrake/gobrake",
        sum = "h1:uTMNQO1LrLNL97C1wh6ZtCZjaWWTkSeOeXyrB0iMQ1s=",
        version = "v3.6.1+incompatible",
    )

    go_repository(
        name = "com_github_alecthomas_template",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/alecthomas/template",
        sum = "h1:JYp7IbQjafoB+tBA3gMyHYHrpOtNuDiK/uB5uXxq5wM=",
        version = "v0.0.0-20190718012654-fb15b899a751",
    )

    go_repository(
        name = "com_github_alecthomas_units",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/alecthomas/units",
        sum = "h1:Hs82Z41s6SdL1CELW+XaDYmOH4hkBN4/N9og/AsOv7E=",
        version = "v0.0.0-20190717042225-c3de453c63f4",
    )

    go_repository(
        name = "com_github_andreyvit_diff",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/andreyvit/diff",
        sum = "h1:bvNMNQO63//z+xNgfBlViaCIJKLlCJ6/fmUseuG0wVQ=",
        version = "v0.0.0-20170406064948-c7f18ee00883",
    )

    go_repository(
        name = "com_github_armon_consul_api",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/armon/consul-api",
        sum = "h1:G1bPvciwNyF7IUmKXNt9Ak3m6u9DE1rF+RmtIkBpVdA=",
        version = "v0.0.0-20180202201655-eb2c6b5be1b6",
    )

    go_repository(
        name = "com_github_asaskevich_govalidator",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/asaskevich/govalidator",
        sum = "h1:zV3ejI06GQ59hwDQAvmK1qxOQGB3WuVTRoY0okPTAv0=",
        version = "v0.0.0-20200108200545-475eaeb16496",
    )

    go_repository(
        name = "com_github_azure_go_ansiterm",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/Azure/go-ansiterm",
        sum = "h1:w+iIsaOQNcT7OZ575w+acHgRric5iCyQh+xv+KJ4HB8=",
        version = "v0.0.0-20170929234023-d6e3b3328b78",
    )

    go_repository(
        name = "com_github_azure_go_autorest_autorest",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/Azure/go-autorest/autorest",
        sum = "h1:NuSF3gXetiHyUbVdneJMEVyPUYAe5wh+aN08JYAf1tI=",
        version = "v0.10.2",
    )

    go_repository(
        name = "com_github_azure_go_autorest_autorest_adal",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/Azure/go-autorest/autorest/adal",
        sum = "h1:Aze/GQeAN1RRbGmnUJvUj+tFGBzFdIg3293/A9rbxC4=",
        version = "v0.9.2",
    )

    go_repository(
        name = "com_github_azure_go_autorest_autorest_date",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/Azure/go-autorest/autorest/date",
        sum = "h1:7gUk1U5M/CQbp9WoqinNzJar+8KY+LPI6wiWrP/myHw=",
        version = "v0.3.0",
    )

    go_repository(
        name = "com_github_azure_go_autorest_autorest_mocks",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/Azure/go-autorest/autorest/mocks",
        sum = "h1:K0laFcLE6VLTOwNgSxaGbUcLPuGXlNkbVvq4cW4nIHk=",
        version = "v0.4.1",
    )

    go_repository(
        name = "com_github_azure_go_autorest_logger",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/Azure/go-autorest/logger",
        sum = "h1:ruG4BSDXONFRrZZJ2GUXDiUyVpayPmb1GnWeHDdaNKY=",
        version = "v0.1.0",
    )

    go_repository(
        name = "com_github_azure_go_autorest_tracing",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/Azure/go-autorest/tracing",
        sum = "h1:TYi4+3m5t6K48TGI9AUdb+IzbnSxvnvUMfuitfgcfuo=",
        version = "v0.6.0",
    )

    go_repository(
        name = "com_github_banzaicloud_k8s_objectmatcher",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/banzaicloud/k8s-objectmatcher",
        sum = "h1:HhXkOWg4xmAW203p3G4Av9ylUfkUJF1+8MK90XQAaJw=",
        version = "v1.3.2",
    )

    go_repository(
        name = "com_github_beorn7_perks",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/beorn7/perks",
        sum = "h1:VlbKKnNfV8bJzeqoa4cOKqO6bYr3WgKZxO8Z16+hsOM=",
        version = "v1.0.1",
    )

    go_repository(
        name = "com_github_bgentry_speakeasy",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/bgentry/speakeasy",
        sum = "h1:ByYyxL9InA1OWqxJqqp2A5pYHUrCiAL6K3J+LKSsQkY=",
        version = "v0.1.0",
    )

    go_repository(
        name = "com_github_bitly_go_simplejson",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/bitly/go-simplejson",
        sum = "h1:6IH+V8/tVMab511d5bn4M7EwGXZf9Hj6i2xSwkNEM+Y=",
        version = "v0.5.0",
    )

    go_repository(
        name = "com_github_blang_semver",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/blang/semver",
        sum = "h1:cQNTCjp13qL8KC3Nbxr/y2Bqb63oX6wdnnjpJbkM4JQ=",
        version = "v3.5.1+incompatible",
    )

    go_repository(
        name = "com_github_bmizerany_assert",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/bmizerany/assert",
        sum = "h1:DDGfHa7BWjL4YnC6+E63dPcxHo2sUxDIu8g3QgEJdRY=",
        version = "v0.0.0-20160611221934-b7ed37b82869",
    )

    go_repository(
        name = "com_github_bugsnag_bugsnag_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/bugsnag/bugsnag-go",
        sum = "h1:CLCt5wO6/P0GelBEMRrlF52XveQMnnXHoCoxGZ+8a5g=",
        version = "v1.4.0",
    )

    go_repository(
        name = "com_github_bugsnag_panicwrap",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/bugsnag/panicwrap",
        sum = "h1:OzrKrRvXis8qEvOkfcxNcYbOd2O7xXS2nnKMEMABFQA=",
        version = "v1.2.0",
    )

    go_repository(
        name = "com_github_burntsushi_toml",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/BurntSushi/toml",
        sum = "h1:WXkYYl6Yr3qBf1K79EBnL4mak0OimBfB0XUf9Vl28OQ=",
        version = "v0.3.1",
    )

    go_repository(
        name = "com_github_burntsushi_xgb",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/BurntSushi/xgb",
        sum = "h1:1BDTz0u9nC3//pOCMdNH+CiXJVYJh5UQNCOBG7jbELc=",
        version = "v0.0.0-20160522181843-27f122750802",
    )

    go_repository(
        name = "com_github_certifi_gocertifi",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/certifi/gocertifi",
        sum = "h1:JLaf/iINcLyjwbtTsCJjc6rtlASgHeIJPrB6QmwURnA=",
        version = "v0.0.0-20200211180108-c7c1fbc02894",
    )

    go_repository(
        name = "com_github_cespare_xxhash_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/cespare/xxhash/v2",
        sum = "h1:6MnRN8NT7+YBpUIWxHtefFZOKTAPgGjpQSxqLNn0+qY=",
        version = "v2.1.1",
    )

    go_repository(
        name = "com_github_client9_misspell",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/client9/misspell",
        sum = "h1:ta993UF76GwbvJcIo3Y68y/M3WxlpEHPWIGDkJYwzJI=",
        version = "v0.3.4",
    )

    go_repository(
        name = "com_github_cockroachdb_datadriven",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/cockroachdb/datadriven",
        sum = "h1:KQqSlwJmTr7NmxtVOl6nPNSZQEk8XuLoLyNDYcmniBo=",
        version = "v1.0.1-0.20200826112548-92602d883b11",
    )

    go_repository(
        name = "com_github_coreos_etcd",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/coreos/etcd",
        sum = "h1:jFneRYjIvLMLhDLCzuTuU4rSJUjRplcJQ7pD7MnhC04=",
        version = "v3.3.10+incompatible",
    )

    go_repository(
        name = "com_github_coreos_go_etcd",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/coreos/go-etcd",
        sum = "h1:bXhRBIXoTm9BYHS3gE0TtQuyNZyeEMux2sDi4oo5YOo=",
        version = "v2.0.0+incompatible",
    )

    go_repository(
        name = "com_github_coreos_go_oidc",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/coreos/go-oidc",
        sum = "h1:mh48q/BqXqgjVHpy2ZY7WnWAbenxRjsz9N1i1YxjHAk=",
        version = "v2.2.1+incompatible",
    )

    go_repository(
        name = "com_github_coreos_go_semver",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/coreos/go-semver",
        sum = "h1:wkHLiw0WNATZnSG7epLsujiMCgPAc9xhjJ4tgnAxmfM=",
        version = "v0.3.0",
    )

    go_repository(
        name = "com_github_coreos_go_systemd",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/coreos/go-systemd",
        sum = "h1:JOrtw2xFKzlg+cbHpyrpLDmnN1HqhBfnX7WDiW7eG2c=",
        version = "v0.0.0-20190719114852-fd7a80b32e1f",
    )

    go_repository(
        name = "com_github_coreos_pkg",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/coreos/pkg",
        sum = "h1:lBNOc5arjvs8E5mO2tbpBpLoyyu8B6e44T7hJy6potg=",
        version = "v0.0.0-20180928190104-399ea9e2e55f",
    )

    go_repository(
        name = "com_github_cpuguy83_go_md2man",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/cpuguy83/go-md2man",
        sum = "h1:BSKMNlYxDvnunlTymqtgONjNnaRV1sTpcovwwjF22jk=",
        version = "v1.0.10",
    )

    go_repository(
        name = "com_github_creack_pty",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/creack/pty",
        sum = "h1:uDmaGzcdjhF4i/plgjmEsriH11Y0o7RKapEf/LDaM3w=",
        version = "v1.1.9",
    )

    go_repository(
        name = "com_github_davecgh_go_spew",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/davecgh/go-spew",
        sum = "h1:vj9j/u1bqnvCEfJOwUhtlOARqs3+rkHYY13jYWTU97c=",
        version = "v1.1.1",
    )

    go_repository(
        name = "com_github_dgrijalva_jwt_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/dgrijalva/jwt-go",
        sum = "h1:7qlOGliEKZXTDg6OTjfoBKDXWrumCAMpl/TFQ4/5kLM=",
        version = "v3.2.0+incompatible",
    )

    go_repository(
        name = "com_github_docker_docker",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/docker/docker",
        sum = "h1:V6pqHzGtZNa2JZXY97Nq+E1IIVwlA90Ql+wDCHPu6Ts=",
        version = "v17.12.0-ce-rc1.0.20190115172544-0dc531243dd3+incompatible",
    )

    go_repository(
        name = "com_github_docker_go_units",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/docker/go-units",
        sum = "h1:3uh0PgVws3nIA0Q+MwDC8yjEPf9zjRfZZWXZYDct3Tw=",
        version = "v0.4.0",
    )

    go_repository(
        name = "com_github_docker_spdystream",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/docker/spdystream",
        sum = "h1:cenwrSVm+Z7QLSV/BsnenAOcDXdX4cMv4wP0B/5QbPg=",
        version = "v0.0.0-20160310174837-449fdfce4d96",
    )

    go_repository(
        name = "com_github_docopt_docopt_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/docopt/docopt-go",
        sum = "h1:bWDMxwH3px2JBh6AyO7hdCn/PkvCZXii8TGj7sbtEbQ=",
        version = "v0.0.0-20180111231733-ee0de3bc6815",
    )

    go_repository(
        name = "com_github_dustin_go_humanize",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/dustin/go-humanize",
        sum = "h1:VSnTsYCnlFHaM2/igO1h6X3HA71jcobQuxemgkq4zYo=",
        version = "v1.0.0",
    )

    go_repository(
        name = "com_github_elazarl_goproxy",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/elazarl/goproxy",
        sum = "h1:yUdfgN0XgIJw7foRItutHYUIhlcKzcSf5vDpdhQAKTc=",
        version = "v0.0.0-20180725130230-947c36da3153",
    )

    go_repository(
        name = "com_github_emicklei_go_restful",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/emicklei/go-restful",
        sum = "h1:spTtZBk5DYEvbxMVutUuTyh1Ao2r4iyvLdACqsl/Ljk=",
        version = "v2.9.5+incompatible",
    )

    go_repository(
        name = "com_github_evanphx_json_patch",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/evanphx/json-patch",
        sum = "h1:ouOWdg56aJriqS0huScTkVXPC5IcNrDCXZ6OoTAWu7M=",
        version = "v4.5.0+incompatible",
    )

    go_repository(
        name = "com_github_fatih_color",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/fatih/color",
        sum = "h1:8xPHl4/q1VyqGIPif1F+1V3Y3lSmrq01EabUW3CoW5s=",
        version = "v1.9.0",
    )

    go_repository(
        name = "com_github_fsnotify_fsnotify",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/fsnotify/fsnotify",
        sum = "h1:hsms1Qyu0jgnwNXIxa+/V/PDsU6CfLf6CNO8H7IWoS4=",
        version = "v1.4.9",
    )

    go_repository(
        name = "com_github_getsentry_raven_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/getsentry/raven-go",
        sum = "h1:no+xWJRb5ZI7eE8TWgIq1jLulQiIoLG0IfYxv5JYMGs=",
        version = "v0.2.0",
    )

    go_repository(
        name = "com_github_ghodss_yaml",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/ghodss/yaml",
        sum = "h1:wQHKEahhL6wmXdzwWG11gIVCkOv05bNOh+Rxn0yngAk=",
        version = "v1.0.0",
    )

    go_repository(
        name = "com_github_globalsign_mgo",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/globalsign/mgo",
        sum = "h1:DujepqpGd1hyOd7aW59XpK7Qymp8iy83xq74fLr21is=",
        version = "v0.0.0-20181015135952-eeefdecb41b8",
    )

    go_repository(
        name = "com_github_go_kit_kit",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/go-kit/kit",
        sum = "h1:wDJmvq38kDhkVxi50ni9ykkdUr1PKgqKOoi01fa0Mdk=",
        version = "v0.9.0",
    )

    go_repository(
        name = "com_github_go_logfmt_logfmt",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/go-logfmt/logfmt",
        sum = "h1:MP4Eh7ZCb31lleYCFuwm0oe4/YGak+5l1vA2NOE80nA=",
        version = "v0.4.0",
    )

    go_repository(
        name = "com_github_go_logr_logr",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/go-logr/logr",
        sum = "h1:M1Tv3VzNlEHg6uyACnRdtrploV2P7wZqH8BoQMtz0cg=",
        version = "v0.1.0",
    )

    go_repository(
        name = "com_github_go_logr_zapr",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/go-logr/zapr",
        sum = "h1:qXBXPDdNncunGs7XeEpsJt8wCjYBygluzfdLO0G5baE=",
        version = "v0.1.1",
    )

    go_repository(
        name = "com_github_go_openapi_analysis",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/go-openapi/analysis",
        sum = "h1:8b2ZgKfKIUTVQpTb77MoRDIMEIwvDVw40o3aOXdfYzI=",
        version = "v0.19.5",
    )

    go_repository(
        name = "com_github_go_openapi_errors",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/go-openapi/errors",
        sum = "h1:a2kIyV3w+OS3S97zxUndRVD46+FhGOUBDFY7nmu4CsY=",
        version = "v0.19.2",
    )

    go_repository(
        name = "com_github_go_openapi_jsonpointer",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/go-openapi/jsonpointer",
        sum = "h1:gihV7YNZK1iK6Tgwwsxo2rJbD1GTbdm72325Bq8FI3w=",
        version = "v0.19.3",
    )

    go_repository(
        name = "com_github_go_openapi_jsonreference",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/go-openapi/jsonreference",
        sum = "h1:5cxNfTy0UVC3X8JL5ymxzyoUZmo8iZb+jeTWn7tUa8o=",
        version = "v0.19.3",
    )

    go_repository(
        name = "com_github_go_openapi_loads",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/go-openapi/loads",
        sum = "h1:5I4CCSqoWzT+82bBkNIvmLc0UOsoKKQ4Fz+3VxOB7SY=",
        version = "v0.19.4",
    )

    go_repository(
        name = "com_github_go_openapi_runtime",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/go-openapi/runtime",
        sum = "h1:csnOgcgAiuGoM/Po7PEpKDoNulCcF3FGbSnbHfxgjMI=",
        version = "v0.19.4",
    )

    go_repository(
        name = "com_github_go_openapi_spec",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/go-openapi/spec",
        sum = "h1:rMMMj8cV38KVXK7SFc+I2MWClbEfbK705+j+dyqun5g=",
        version = "v0.19.6",
    )

    go_repository(
        name = "com_github_go_openapi_strfmt",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/go-openapi/strfmt",
        sum = "h1:eRfyY5SkaNJCAwmmMcADjY31ow9+N7MCLW7oRkbsINA=",
        version = "v0.19.3",
    )

    go_repository(
        name = "com_github_go_openapi_swag",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/go-openapi/swag",
        sum = "h1:VRuXN2EnMSsZdauzdss6JBC29YotDqG59BZ+tdlIL1s=",
        version = "v0.19.7",
    )

    go_repository(
        name = "com_github_go_openapi_validate",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/go-openapi/validate",
        sum = "h1:QhCBKRYqZR+SKo4gl1lPhPahope8/RLt6EVgY8X80w0=",
        version = "v0.19.5",
    )

    go_repository(
        name = "com_github_go_stack_stack",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/go-stack/stack",
        sum = "h1:5SgMzNM5HxrEjV0ww2lTmX6E2Izsfxas4+YHWRs3Lsk=",
        version = "v1.8.0",
    )

    go_repository(
        name = "com_github_gofrs_uuid",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/gofrs/uuid",
        sum = "h1:8K4tyRfvU1CYPgJsveYFQMhpFd/wXNM7iK6rR7UHz84=",
        version = "v3.3.0+incompatible",
    )

    go_repository(
        name = "com_github_gogo_protobuf",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/gogo/protobuf",
        replace = "github.com/cockroachdb/gogoproto",
        sum = "h1:yrdrJWJpn0+1BmXaRzurDIZz3uHmSs/wnwWbC4arWlQ=",
        version = "v1.2.1-0.20210111172841-8b6737fea948",
    )

    go_repository(
        name = "com_github_golang_glog",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/golang/glog",
        sum = "h1:VKtxabqXZkF25pY9ekfRL6a582T4P37/31XEstQ5p58=",
        version = "v0.0.0-20160126235308-23def4e6c14b",
    )

    go_repository(
        name = "com_github_golang_groupcache",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/golang/groupcache",
        sum = "h1:1r7pUrabqp18hOBcwBwiTsbnFeTZHV9eER/QT5JVZxY=",
        version = "v0.0.0-20200121045136-8c9f03a8e57e",
    )

    go_repository(
        name = "com_github_golang_mock",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/golang/mock",
        sum = "h1:qGJ6qTW+x6xX/my+8YUVl4WNpX9B7+/l2tRsHGZ7f2s=",
        version = "v1.3.1",
    )

    go_repository(
        name = "com_github_golang_protobuf",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/golang/protobuf",
        sum = "h1:+Z5KGCizgyZCbGh1KZqA0fcLLkwbsjIzS4aV2v7wJX0=",
        version = "v1.4.2",
    )

    go_repository(
        name = "com_github_google_btree",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/google/btree",
        sum = "h1:0udJVsspx3VBr5FwtLhQQtuAsVc79tTq0ocGIPAU6qo=",
        version = "v1.0.0",
    )

    go_repository(
        name = "com_github_google_go_cmp",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/google/go-cmp",
        sum = "h1:xsAVV57WRhGj6kEIi8ReJzQlHHqcBYCElAvkovg3B/4=",
        version = "v0.4.0",
    )

    go_repository(
        name = "com_github_google_gofuzz",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/google/gofuzz",
        sum = "h1:Hsa8mG0dQ46ij8Sl2AYJDUv1oA9/d6Vk+3LG99Oe02g=",
        version = "v1.1.0",
    )

    go_repository(
        name = "com_github_google_martian",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/google/martian",
        sum = "h1:xmapqc1AyLoB+ddYT6r04bD9lIjlOqGaREovi0SzFaE=",
        version = "v2.1.1-0.20190517191504-25dcb96d9e51+incompatible",
    )

    go_repository(
        name = "com_github_google_pprof",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/google/pprof",
        sum = "h1:XTnP8fJpa4Kvpw2qARB4KS9izqxPS0Sd92cDlY3uk+w=",
        version = "v0.0.0-20190723021845-34ac40c74b70",
    )

    go_repository(
        name = "com_github_google_uuid",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/google/uuid",
        sum = "h1:Gkbcsh/GbpXz7lPftLA3P6TYMwjCLYm83jiFQZF/3gY=",
        version = "v1.1.1",
    )

    go_repository(
        name = "com_github_googleapis_gax_go_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/googleapis/gax-go/v2",
        sum = "h1:sjZBwGj9Jlw33ImPtvFviGYvseOtDM7hkSKB7+Tv3SM=",
        version = "v2.0.5",
    )

    go_repository(
        name = "com_github_googleapis_gnostic",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/googleapis/gnostic",
        sum = "h1:WeAefnSUHlBb0iJKwxFDZdbfGwkd7xRNuV+IpXMJhYk=",
        version = "v0.3.1",
    )

    go_repository(
        name = "com_github_goph_emperror",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/goph/emperror",
        sum = "h1:yLapQcmEsO0ipe9p5TaN22djm3OFV/TfM/fcYP0/J18=",
        version = "v0.17.2",
    )

    go_repository(
        name = "com_github_gophercloud_gophercloud",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/gophercloud/gophercloud",
        sum = "h1:P/nh25+rzXouhytV2pUHBb65fnds26Ghl8/391+sT5o=",
        version = "v0.1.0",
    )

    go_repository(
        name = "com_github_gorilla_websocket",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/gorilla/websocket",
        sum = "h1:WDFjx/TMzVgy9VdMMQi2K2Emtwi2QcUQsztZ/zLaH/Q=",
        version = "v1.4.0",
    )

    go_repository(
        name = "com_github_gregjones_httpcache",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/gregjones/httpcache",
        sum = "h1:f8eY6cV/x1x+HLjOp4r72s/31/V2aTUtg5oKRRPf8/Q=",
        version = "v0.0.0-20190212212710-3befbb6ad0cc",
    )

    go_repository(
        name = "com_github_grpc_ecosystem_go_grpc_middleware",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/grpc-ecosystem/go-grpc-middleware",
        sum = "h1:z53tR0945TRRQO/fLEVPI6SMv7ZflF0TEaTAoU7tOzg=",
        version = "v1.0.1-0.20190118093823-f849b5445de4",
    )

    go_repository(
        name = "com_github_grpc_ecosystem_go_grpc_prometheus",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/grpc-ecosystem/go-grpc-prometheus",
        sum = "h1:Ovs26xHkKqVztRpIrF/92BcuyuQ/YW4NSIpoGtfXNho=",
        version = "v1.2.0",
    )

    go_repository(
        name = "com_github_grpc_ecosystem_grpc_gateway",
        build_file_generation = "on",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/grpc-ecosystem/grpc-gateway",
        replace = "github.com/cockroachdb/grpc-gateway",
        sum = "h1:pZuUpNtmLArGU2vCW+mNetvruw6jRcxDxwOnMT+XLWI=",
        version = "v1.14.6-0.20200519165156-52697fc4a249",
    )

    go_repository(
        name = "com_github_hashicorp_golang_lru",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/hashicorp/golang-lru",
        sum = "h1:YDjusn29QI/Das2iO9M0BHnIbxPeyuCHsjMW+lJfyTc=",
        version = "v0.5.4",
    )

    go_repository(
        name = "com_github_hashicorp_hcl",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/hashicorp/hcl",
        sum = "h1:0Anlzjpi4vEasTeNFn2mLJgTSwt0+6sfsiTG8qcWGx4=",
        version = "v1.0.0",
    )

    go_repository(
        name = "com_github_hpcloud_tail",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/hpcloud/tail",
        sum = "h1:nfCOvKYfkgYP8hkirhJocXT2+zOD8yUNjXaWfTlyFKI=",
        version = "v1.0.0",
    )

    go_repository(
        name = "com_github_imdario_mergo",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/imdario/mergo",
        sum = "h1:UauaLniWCFHWd+Jp9oCEkTBj8VO/9DKg3PV3VCNMDIg=",
        version = "v0.3.9",
    )

    go_repository(
        name = "com_github_inconshreveable_mousetrap",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/inconshreveable/mousetrap",
        sum = "h1:Z8tu5sraLXCXIcARxBp/8cbvlwVa7Z1NHg9XEKhtSvM=",
        version = "v1.0.0",
    )

    go_repository(
        name = "com_github_jonboulle_clockwork",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/jonboulle/clockwork",
        sum = "h1:VKV+ZcuP6l3yW9doeqz6ziZGgcynBVQO+obU0+0hcPo=",
        version = "v0.1.0",
    )

    go_repository(
        name = "com_github_json_iterator_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/json-iterator/go",
        sum = "h1:9yzud/Ht36ygwatGx56VwCZtlI/2AD15T1X2sjSuGns=",
        version = "v1.1.9",
    )

    go_repository(
        name = "com_github_jstemmer_go_junit_report",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/jstemmer/go-junit-report",
        sum = "h1:rBMNdlhTLzJjJSDIjNEXX1Pz3Hmwmz91v+zycvx9PJc=",
        version = "v0.0.0-20190106144839-af01ea7f8024",
    )

    go_repository(
        name = "com_github_julienschmidt_httprouter",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/julienschmidt/httprouter",
        sum = "h1:TDTW5Yz1mjftljbcKqRcrYhd4XeOoI98t+9HbQbYf7g=",
        version = "v1.2.0",
    )

    go_repository(
        name = "com_github_kardianos_osext",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/kardianos/osext",
        sum = "h1:iQTw/8FWTuc7uiaSepXwyf3o52HaUYcV+Tu66S3F5GA=",
        version = "v0.0.0-20190222173326-2bc1f35cddc0",
    )

    go_repository(
        name = "com_github_kisielk_errcheck",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/kisielk/errcheck",
        sum = "h1:reN85Pxc5larApoH1keMBiu2GWtPqXQ1nc9gx+jOU+E=",
        version = "v1.2.0",
    )

    go_repository(
        name = "com_github_kisielk_gotool",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/kisielk/gotool",
        sum = "h1:AV2c/EiW3KqPNT9ZKl07ehoAGi4C5/01Cfbblndcapg=",
        version = "v1.0.0",
    )

    go_repository(
        name = "com_github_konsorten_go_windows_terminal_sequences",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/konsorten/go-windows-terminal-sequences",
        sum = "h1:CE8S1cTafDpPvMhIxNJKvHsGVBgn1xWYf1NbHQhywc8=",
        version = "v1.0.3",
    )

    go_repository(
        name = "com_github_kr_logfmt",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/kr/logfmt",
        sum = "h1:T+h1c/A9Gawja4Y9mFVWj2vyii2bbUNDw3kt9VxK2EY=",
        version = "v0.0.0-20140226030751-b84e30acd515",
    )

    go_repository(
        name = "com_github_kr_pretty",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/kr/pretty",
        sum = "h1:Fmg33tUaq4/8ym9TJN1x7sLJnHVwhP33CNkpYV/7rwI=",
        version = "v0.2.1",
    )

    go_repository(
        name = "com_github_kr_pty",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/kr/pty",
        sum = "h1:AkaSdXYQOWeaO3neb8EM634ahkXXe3jYbVh/F9lq+GI=",
        version = "v1.1.8",
    )

    go_repository(
        name = "com_github_kr_text",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/kr/text",
        sum = "h1:5Nx0Ya0ZqY2ygV366QzturHI13Jq95ApcVaJBhpS+AY=",
        version = "v0.2.0",
    )

    go_repository(
        name = "com_github_magiconair_properties",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/magiconair/properties",
        sum = "h1:ZC2Vc7/ZFkGmsVC9KvOjumD+G5lXy2RtTKyzRKO2BQ4=",
        version = "v1.8.1",
    )

    go_repository(
        name = "com_github_mailru_easyjson",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/mailru/easyjson",
        sum = "h1:aizVhC/NAAcKWb+5QsU1iNOZb4Yws5UO2I+aIprQITM=",
        version = "v0.7.0",
    )

    go_repository(
        name = "com_github_masterminds_semver_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/Masterminds/semver/v3",
        sum = "h1:Y2lUDsFKVRSYGojLJ1yLxSXdMmMYTYls0rCvoqmMUQk=",
        version = "v3.1.0",
    )

    go_repository(
        name = "com_github_mattn_go_colorable",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/mattn/go-colorable",
        sum = "h1:6Su7aK7lXmJ/U79bYtBjLNaha4Fs1Rg9plHpcH+vvnE=",
        version = "v0.1.6",
    )

    go_repository(
        name = "com_github_mattn_go_isatty",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/mattn/go-isatty",
        sum = "h1:wuysRhFDzyxgEmMf5xjvJ2M9dZoWAXNNr5LSBS7uHXY=",
        version = "v0.0.12",
    )

    go_repository(
        name = "com_github_mattn_go_runewidth",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/mattn/go-runewidth",
        sum = "h1:3tS41NlGYSmhhe/8fhGRzc+z3AYCw1Fe1WAyLuujKs0=",
        version = "v0.0.8",
    )

    go_repository(
        name = "com_github_matttproud_golang_protobuf_extensions",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/matttproud/golang_protobuf_extensions",
        sum = "h1:4hp9jkHxhMHkqkrB3Ix0jegS5sx/RkqARlsWZ6pIwiU=",
        version = "v1.0.1",
    )

    go_repository(
        name = "com_github_mitchellh_go_homedir",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/mitchellh/go-homedir",
        sum = "h1:lukF9ziXFxDFPkA1vsr5zpc1XuPDn/wFntq5mG+4E0Y=",
        version = "v1.1.0",
    )

    go_repository(
        name = "com_github_mitchellh_mapstructure",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/mitchellh/mapstructure",
        sum = "h1:fmNYVwqnSfB9mZU6OS2O6GsXM+wcskZDuKQzvN1EDeE=",
        version = "v1.1.2",
    )

    go_repository(
        name = "com_github_modern_go_concurrent",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/modern-go/concurrent",
        sum = "h1:TRLaZ9cD/w8PVh93nsPXa1VrQ6jlwL5oN8l14QlcNfg=",
        version = "v0.0.0-20180306012644-bacd9c7ef1dd",
    )

    go_repository(
        name = "com_github_modern_go_reflect2",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/modern-go/reflect2",
        sum = "h1:9f412s+6RmYXLWZSEzVVgPGK7C2PphHj5RJrvfx9AWI=",
        version = "v1.0.1",
    )

    go_repository(
        name = "com_github_munnerz_goautoneg",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/munnerz/goautoneg",
        sum = "h1:C3w9PqII01/Oq1c1nUAm88MOHcQC9l5mIlSMApZMrHA=",
        version = "v0.0.0-20191010083416-a7dc8b61c822",
    )

    go_repository(
        name = "com_github_mwitkow_go_conntrack",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/mwitkow/go-conntrack",
        sum = "h1:KUppIJq7/+SVif2QVs3tOP0zanoHgBEVAwHxUSIzRqU=",
        version = "v0.0.0-20190716064945-2f068394615f",
    )

    go_repository(
        name = "com_github_mxk_go_flowrate",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/mxk/go-flowrate",
        sum = "h1:y5//uYreIhSUg3J1GEMiLbxo1LJaP8RfCpH6pymGZus=",
        version = "v0.0.0-20140419014527-cca7078d478f",
    )

    go_repository(
        name = "com_github_nytimes_gziphandler",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/NYTimes/gziphandler",
        sum = "h1:lsxEuwrXEAokXB9qhlbKWPpo3KMLZQ5WB5WLQRW1uq0=",
        version = "v0.0.0-20170623195520-56545f4a5d46",
    )

    go_repository(
        name = "com_github_olekukonko_tablewriter",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/olekukonko/tablewriter",
        replace = "github.com/cockroachdb/tablewriter",
        sum = "h1:c7yLgqcm/3c9lYtpWeVD9NYqA9cKsKHdpQM62PHtTUM=",
        version = "v0.0.5-0.20200105123400-bd15540e8847",
    )

    go_repository(
        name = "com_github_onsi_ginkgo",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/onsi/ginkgo",
        sum = "h1:M76yO2HkZASFjXL0HSoZJ1AYEmQxNJmY41Jx1zNUq1Y=",
        version = "v1.13.0",
    )

    go_repository(
        name = "com_github_onsi_gomega",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/onsi/gomega",
        sum = "h1:o0+MgICZLuZ7xjH7Vx6zS/zcu93/BEp1VwkIW1mEXCE=",
        version = "v1.10.1",
    )

    go_repository(
        name = "com_github_pborman_uuid",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/pborman/uuid",
        sum = "h1:J7Q5mO4ysT1dv8hyrUGHb9+ooztCXu1D8MY8DZYsu3g=",
        version = "v1.2.0",
    )

    go_repository(
        name = "com_github_pelletier_go_toml",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/pelletier/go-toml",
        sum = "h1:aetoXYr0Tv7xRU/V4B4IZJ2QcbtMUFoNb3ORp7TzIK4=",
        version = "v1.6.0",
    )

    go_repository(
        name = "com_github_peterbourgon_diskv",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/peterbourgon/diskv",
        sum = "h1:UBdAOUP5p4RWqPBg048CAvpKN+vxiaj6gdUUzhl4XmI=",
        version = "v2.0.1+incompatible",
    )

    go_repository(
        name = "com_github_pkg_errors",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/pkg/errors",
        sum = "h1:FEBLx1zS214owpjy7qsBeixbURkuhQAwrK5UwLGTwt4=",
        version = "v0.9.1",
    )

    go_repository(
        name = "com_github_pmezard_go_difflib",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/pmezard/go-difflib",
        sum = "h1:4DBwDE0NGyQoBHbLQYPwSUPoCMWR5BEzIk/f1lZbAQM=",
        version = "v1.0.0",
    )

    go_repository(
        name = "com_github_pquerna_cachecontrol",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/pquerna/cachecontrol",
        sum = "h1:JDEmUDtyiLMyMlFwiaDOv2hxUp35497fkwePcLeV7j4=",
        version = "v0.0.0-20200819021114-67c6ae64274f",
    )

    go_repository(
        name = "com_github_prometheus_client_golang",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/prometheus/client_golang",
        sum = "h1:Ctq0iGpCmr3jeP77kbF2UxgvRwzWWz+4Bh9/vJTyg1A=",
        version = "v1.5.0",
    )

    go_repository(
        name = "com_github_prometheus_client_model",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/prometheus/client_model",
        sum = "h1:uq5h0d+GuxiXLJLNABMgp2qUWDPiLvgCzz2dUR+/W/M=",
        version = "v0.2.0",
    )

    go_repository(
        name = "com_github_prometheus_common",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/prometheus/common",
        sum = "h1:KOMtN28tlbam3/7ZKEYKHhKoJZYYj3gMH4uc62x7X7U=",
        version = "v0.9.1",
    )

    go_repository(
        name = "com_github_prometheus_procfs",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/prometheus/procfs",
        sum = "h1:QJQN3jYQhkamO4mhfUWqdDH2asK7ONOI9MTWjyAxNKM=",
        version = "v0.0.10",
    )

    go_repository(
        name = "com_github_puerkitobio_purell",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/PuerkitoBio/purell",
        sum = "h1:WEQqlqaGbrPkxLJWfBwQmfEAE1Z7ONdDLqrN38tNFfI=",
        version = "v1.1.1",
    )

    go_repository(
        name = "com_github_puerkitobio_urlesc",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/PuerkitoBio/urlesc",
        sum = "h1:d+Bc7a5rLufV/sSk/8dngufqelfh6jnri85riMAaF/M=",
        version = "v0.0.0-20170810143723-de5bf2ad4578",
    )

    go_repository(
        name = "com_github_remyoudompheng_bigfft",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/remyoudompheng/bigfft",
        sum = "h1:/NRJ5vAYoqz+7sG51ubIDHXeWO8DlTSrToPu6q11ziA=",
        version = "v0.0.0-20170806203942-52369c62f446",
    )

    go_repository(
        name = "com_github_rogpeppe_fastuuid",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/rogpeppe/fastuuid",
        sum = "h1:Ppwyp6VYCF1nvBTXL3trRso7mXMlRrw9ooo375wvi2s=",
        version = "v1.2.0",
    )

    go_repository(
        name = "com_github_rollbar_rollbar_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/rollbar/rollbar-go",
        sum = "h1:uA3+z0jq6ka9WUUt9VX/xuiQZXZyWRoeKvkhVvLO9Jc=",
        version = "v1.0.2",
    )

    go_repository(
        name = "com_github_russross_blackfriday",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/russross/blackfriday",
        sum = "h1:HyvC0ARfnZBqnXwABFeSZHpKvJHJJfPz81GNueLj0oo=",
        version = "v1.5.2",
    )

    go_repository(
        name = "com_github_sergi_go_diff",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/sergi/go-diff",
        sum = "h1:we8PVUC3FE2uYfodKH/nBHMSetSfHDR6scGdBi+erh0=",
        version = "v1.1.0",
    )

    go_repository(
        name = "com_github_sirupsen_logrus",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/sirupsen/logrus",
        sum = "h1:UBcNElsrwanuuMsnGSlYmtmgbb23qDR5dG+6X6Oo89I=",
        version = "v1.6.0",
    )

    go_repository(
        name = "com_github_soheilhy_cmux",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/soheilhy/cmux",
        sum = "h1:0HKaf1o97UwFjHH9o5XsHUOF+tqmdA7KEzXLpiyaw0E=",
        version = "v0.1.4",
    )

    go_repository(
        name = "com_github_spf13_afero",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/spf13/afero",
        sum = "h1:5jhuqJyZCZf2JRofRvN/nIFgIWNzPa3/Vz8mYylgbWc=",
        version = "v1.2.2",
    )

    go_repository(
        name = "com_github_spf13_cast",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/spf13/cast",
        sum = "h1:nFm6S0SMdyzrzcmThSipiEubIDy8WEXKNZ0UOgiRpng=",
        version = "v1.3.1",
    )

    go_repository(
        name = "com_github_spf13_cobra",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/spf13/cobra",
        sum = "h1:6m/oheQuQ13N9ks4hubMG6BnvwOeaJrqSPLahSnczz8=",
        version = "v1.0.0",
    )

    go_repository(
        name = "com_github_spf13_jwalterweatherman",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/spf13/jwalterweatherman",
        sum = "h1:ue6voC5bR5F8YxI5S67j9i582FU4Qvo2bmqnqMYADFk=",
        version = "v1.1.0",
    )

    go_repository(
        name = "com_github_spf13_pflag",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/spf13/pflag",
        sum = "h1:iy+VFUOCP1a+8yFto/drg2CJ5u0yRoB7fZw3DKv/JXA=",
        version = "v1.0.5",
    )

    go_repository(
        name = "com_github_spf13_viper",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/spf13/viper",
        sum = "h1:7aKfF+e8/k68gda3LOjo5RxiUqddoFxVq4BKBPrxk5E=",
        version = "v1.6.2",
    )

    go_repository(
        name = "com_github_stretchr_objx",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/stretchr/objx",
        sum = "h1:Hbg2NidpLE8veEBkEZTL3CvlkUIVzuU9jDplZO54c48=",
        version = "v0.2.0",
    )

    go_repository(
        name = "com_github_stretchr_testify",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/stretchr/testify",
        sum = "h1:hDPOHmpOpP40lSULcqw7IrRb/u7w6RpDC9399XyoNd0=",
        version = "v1.6.1",
    )

    go_repository(
        name = "com_github_tidwall_pretty",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/tidwall/pretty",
        sum = "h1:HsD+QiTn7sK6flMKIvNmpqz1qrpP3Ps6jOKIKMooyg4=",
        version = "v1.0.0",
    )

    go_repository(
        name = "com_github_tmc_grpc_websocket_proxy",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/tmc/grpc-websocket-proxy",
        sum = "h1:LnC5Kc/wtumK+WB441p7ynQJzVuNRJiqddSIE3IlSEQ=",
        version = "v0.0.0-20190109142713-0ad062ec5ee5",
    )

    go_repository(
        name = "com_github_ugorji_go_codec",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/ugorji/go/codec",
        sum = "h1:3SVOIvH7Ae1KRYyQWRjXWJEA9sS/c/pjvH++55Gr648=",
        version = "v0.0.0-20181204163529-d75b2dcb6bc8",
    )

    go_repository(
        name = "com_github_urfave_cli",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/urfave/cli",
        sum = "h1:fDqGv3UG/4jbVl/QkFwEdddtEDjh/5Ov6X+0B/3bPaw=",
        version = "v1.20.0",
    )

    go_repository(
        name = "com_github_vektah_gqlparser",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/vektah/gqlparser",
        sum = "h1:ZsyLGn7/7jDNI+y4SEhI4yAxRChlv15pUHMjijT+e68=",
        version = "v1.1.2",
    )

    go_repository(
        name = "com_github_xiang90_probing",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/xiang90/probing",
        sum = "h1:eY9dn8+vbi4tKz5Qo6v2eYzo7kUS51QINcR5jNpbZS8=",
        version = "v0.0.0-20190116061207-43a291ad63a2",
    )

    go_repository(
        name = "com_github_xordataexchange_crypt",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/xordataexchange/crypt",
        sum = "h1:ESFSdwYZvkeru3RtdrYueztKhOBCSAAzS4Gf+k0tEow=",
        version = "v0.0.3-0.20170626215501-b2862e3d0a77",
    )

    go_repository(
        name = "com_google_cloud_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "cloud.google.com/go",
        sum = "h1:1JUtpcY9E7+eTospEwWS2QXP3DEn7poB3E2j0jN74mM=",
        version = "v0.47.0",
    )

    go_repository(
        name = "in_gopkg_alecthomas_kingpin_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "gopkg.in/alecthomas/kingpin.v2",
        sum = "h1:jMFz6MfLP0/4fUyZle81rXUoxOBFi19VUFKVDOQfozc=",
        version = "v2.2.6",
    )

    go_repository(
        name = "in_gopkg_check_v1",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "gopkg.in/check.v1",
        sum = "h1:BLraFXnmrev5lT+xlilqcH8XK9/i0At2xKjWk4p6zsU=",
        version = "v1.0.0-20200227125254-8fa46927fb4f",
    )

    go_repository(
        name = "in_gopkg_cheggaaa_pb_v1",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "gopkg.in/cheggaaa/pb.v1",
        sum = "h1:Ev7yu1/f6+d+b3pi5vPdRPc6nNtP1umSfcWiEfRqv6I=",
        version = "v1.0.25",
    )

    go_repository(
        name = "in_gopkg_fsnotify_v1",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "gopkg.in/fsnotify.v1",
        sum = "h1:xOHLXZwVvI9hhs+cLKq5+I5onOuwQLhQwiu63xxlHs4=",
        version = "v1.4.7",
    )

    go_repository(
        name = "in_gopkg_inf_v0",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "gopkg.in/inf.v0",
        sum = "h1:73M5CoZyi3ZLMOyDlQh031Cx6N9NDJ2Vvfl76EDAgDc=",
        version = "v0.9.1",
    )

    go_repository(
        name = "in_gopkg_natefinch_lumberjack_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "gopkg.in/natefinch/lumberjack.v2",
        sum = "h1:1Lc07Kr7qY4U2YPouBjpCLxpiyxIVoxqXgkXLknAOE8=",
        version = "v2.0.0",
    )

    go_repository(
        name = "in_gopkg_resty_v1",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "gopkg.in/resty.v1",
        sum = "h1:CuXP0Pjfw9rOuY6EP+UvtNvt5DSqHpIxILZKT/quCZI=",
        version = "v1.12.0",
    )

    go_repository(
        name = "in_gopkg_square_go_jose_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "gopkg.in/square/go-jose.v2",
        sum = "h1:7odma5RETjNHWJnR32wx8t+Io4djHE1PqxCFx3iiZ2w=",
        version = "v2.5.1",
    )

    go_repository(
        name = "in_gopkg_tomb_v1",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "gopkg.in/tomb.v1",
        sum = "h1:uRGJdciOHaEIrze2W8Q3AKkepLTh2hOroT7a+7czfdQ=",
        version = "v1.0.0-20141024135613-dd632973f1e7",
    )

    go_repository(
        name = "in_gopkg_yaml_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "gopkg.in/yaml.v2",
        replace = "github.com/cockroachdb/yaml",
        sum = "h1:EqoCicA1pbWWDGniFxhTElh2hvui7E7tEvuBNJSDn3A=",
        version = "v0.0.0-20180705215940-0e2822948641",
    )

    go_repository(
        name = "io_etcd_go_bbolt",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "go.etcd.io/bbolt",
        sum = "h1:MUGmc65QhB3pIlaQ5bB4LwqSj6GIonVJXpZiaKNyaKk=",
        version = "v1.3.3",
    )

    go_repository(
        name = "io_etcd_go_etcd",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "go.etcd.io/etcd",
        replace = "github.com/cockroachdb/etcd",
        sum = "h1:1ELogBFqZl2Cj/Rn0wWAU4z2/ghBlHzhXDlbT8RlrKU=",
        version = "v0.4.7-0.20200615211340-a17df30d5955",
    )
    go_repository(
        name = "io_k8s_api",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "k8s.io/api",
        sum = "h1:fKbCxr+U3fu7k6jB+QeYPD/c6xKYeSJ2KVWmyUypuWM=",
        version = "v0.18.5",
    )

    go_repository(
        name = "io_k8s_apiextensions_apiserver",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "k8s.io/apiextensions-apiserver",
        sum = "h1:pvbXjB/BRXZiO+/Erp5Pxr+lnhDCv5uxNxHh3FLGZ/g=",
        version = "v0.18.5",
    )

    go_repository(
        name = "io_k8s_apimachinery",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "k8s.io/apimachinery",
        sum = "h1:Lh6tgsM9FMkC12K5T5QjRm7rDs6aQN5JHkA0JomULDM=",
        version = "v0.18.5",
    )

    go_repository(
        name = "io_k8s_apiserver",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "k8s.io/apiserver",
        sum = "h1:+kk0J2Qf6xd8a5rxcNpf80jwCF99KUWA7tpvctTWObc=",
        version = "v0.18.5",
    )

    go_repository(
        name = "io_k8s_client_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "k8s.io/client-go",
        replace = "k8s.io/client-go",
        sum = "h1:cLhGZdOmyPhwtt20Lrb7uAqxxB1uvY+NTmNJvno1oKA=",
        version = "v0.18.5",
    )

    go_repository(
        name = "io_k8s_code_generator",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "k8s.io/code-generator",
        sum = "h1:qMh1fOcU/jOe52e/Sc0ZAnicSk1TCxx81foIYuBIzGk=",
        version = "v0.18.5",
    )

    go_repository(
        name = "io_k8s_component_base",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "k8s.io/component-base",
        sum = "h1:QIz2xFax8W5XQREoNcP/MHgnt4ClgfZ837Qx9yCeCzA=",
        version = "v0.18.5",
    )

    go_repository(
        name = "io_k8s_gengo",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "k8s.io/gengo",
        sum = "h1:RPscN6KhmG54S33L+lr3GS+oD1jmchIU0ll519K6FA4=",
        version = "v0.0.0-20200114144118-36b2048a9120",
    )

    go_repository(
        name = "io_k8s_klog",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "k8s.io/klog",
        sum = "h1:Pt+yjF5aB1xDSVbau4VsWe+dQNzA0qv1LlXdC2dF6Q8=",
        version = "v1.0.0",
    )

    go_repository(
        name = "io_k8s_kube_aggregator",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "k8s.io/kube-aggregator",
        sum = "h1:mGRE5L7cm2Uucp5LNxwEcf60apGotk/g0+awoZ0JnGg=",
        version = "v0.18.5",
    )

    go_repository(
        name = "io_k8s_kube_openapi",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "k8s.io/kube-openapi",
        sum = "h1:Oh3Mzx5pJ+yIumsAD0MOECPVeXsVot0UkiaCGVyfGQY=",
        version = "v0.0.0-20200410145947-61e04a5be9a6",
    )

    go_repository(
        name = "io_k8s_sigs_controller_runtime",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "sigs.k8s.io/controller-runtime",
        replace = "sigs.k8s.io/controller-runtime",
        sum = "h1:J1LXpXsQccTWk2V2OZZBc4mC2tEL70bBb/cXQyVhOYY=",
        version = "v0.5.1-0.20200416234307-5377effd4043",
    )

    #FIXME this is not updating via gazelle automatically
    go_repository(
        name = "io_k8s_sigs_controller_tools",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "sigs.k8s.io/controller-tools",
        sum = "h1:PXOHvyYAjWfO0UfQvaUo33HpXNCOilV3i/Vjc7iM1/A=",
        version = "v0.2.9-0.20200414181213-645d44dca7c0",
    )

    go_repository(
        name = "io_k8s_apimachinery",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "k8s.io/apimachinery",
        sum = "h1:Lh6tgsM9FMkC12K5T5QjRm7rDs6aQN5JHkA0JomULDM=",
        version = "v0.18.5",
    )

    go_repository(
        name = "in_gopkg_yaml_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "gopkg.in/yaml.v3",
        sum = "h1:dUUwHk2QECo/6vqA44rthZ8ie2QXMNeKRTHCNY2nXvo=",
        version = "v3.0.0-20200313102051-9f266ea9e77c",
    )

    go_repository(
        name = "com_github_gobuffalo_flect",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/gobuffalo/flect",
        sum = "h1:EWCvMGGxOjsgwlWaP+f4+Hh6yrrte7JeFL2S6b+0hdM=",
        version = "v0.2.0",
    )

    # end manual updates

    go_repository(
        name = "io_k8s_sigs_structured_merge_diff",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "sigs.k8s.io/structured-merge-diff",
        sum = "h1:zD2IemQ4LmOcAumeiyDWXKUI2SO0NYDe3H6QGvPOVgU=",
        version = "v1.0.1-0.20191108220359-b1b620dd3f06",
    )

    go_repository(
        name = "io_k8s_sigs_structured_merge_diff_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "sigs.k8s.io/structured-merge-diff/v3",
        sum = "h1:dOmIZBMfhcHS09XZkMyUgkq5trg3/jRyJYFZUiaOp8E=",
        version = "v3.0.0",
    )

    go_repository(
        name = "io_k8s_sigs_yaml",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "sigs.k8s.io/yaml",
        sum = "h1:kr/MCeFWJWTwyaHoR9c8EjH9OumOmoF9YGiZd7lFm/Q=",
        version = "v1.2.0",
    )

    go_repository(
        name = "io_k8s_utils",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "k8s.io/utils",
        sum = "h1:d4vVOjXm687F1iLSP2q3lyPPuyvTUt3aVoBpi2DqRsU=",
        version = "v0.0.0-20200324210504-a9aa75ae1b89",
    )

    go_repository(
        name = "io_opencensus_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "go.opencensus.io",
        sum = "h1:8sGtKOrtQqkN1bp2AtX+misvLIlOmsEsNd+9NIcPEm8=",
        version = "v0.22.3",
    )

    go_repository(
        name = "org_golang_google_api",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "google.golang.org/api",
        sum = "h1:yzlyyDW/J0w8yNFJIhiAJy4kq74S+1DOLdawELNxFMA=",
        version = "v0.15.0",
    )

    go_repository(
        name = "org_golang_google_appengine",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "google.golang.org/appengine",
        sum = "h1:tycE03LOZYQNhDpS27tcQdAzLCVMaj7QT2SXxebnpCM=",
        version = "v1.6.5",
    )

    go_repository(
        name = "org_golang_google_genproto",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "google.golang.org/genproto",
        sum = "h1:jB9+PJSvu5tBfmJHy/OVapFdjDF3WvpkqRhxqrmzoEU=",
        version = "v0.0.0-20200218151345-dad8c97a84f5",
    )

    go_repository(
        name = "org_golang_google_grpc",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "google.golang.org/grpc",
        sum = "h1:EC2SB8S04d2r73uptxphDSUG+kTKVgjRPF+N3xpxRB4=",
        version = "v1.29.1",
    )

    go_repository(
        name = "org_golang_x_crypto",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "golang.org/x/crypto",
        sum = "h1:psW17arqaxU48Z5kZ0CQnkZWQJsqcURM6tKiBApRjXI=",
        version = "v0.0.0-20200622213623-75b288015ac9",
    )

    go_repository(
        name = "org_golang_x_exp",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "golang.org/x/exp",
        sum = "h1:i1+omYRtqpxiCaQJB4MQhUToKvMPFqUUJKvRiRp0gtE=",
        version = "v0.0.0-20200901203048-c4f52b2c50aa",
    )

    go_repository(
        name = "org_golang_x_image",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "golang.org/x/image",
        sum = "h1:+qEpEAPhDZ1o0x3tHzZTQDArnOixOzGD9HUJfcg0mb4=",
        version = "v0.0.0-20190802002840-cff245a6509b",
    )

    go_repository(
        name = "org_golang_x_lint",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "golang.org/x/lint",
        sum = "h1:Wh+f8QHJXR411sJR8/vRBTZ7YapZaRvUcLFFJhusH0k=",
        version = "v0.0.0-20200302205851-738671d3881b",
    )

    go_repository(
        name = "org_golang_x_mobile",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "golang.org/x/mobile",
        sum = "h1:b373EGXtj0o+ssqkOkdVphTCZ/fVg2LwhctJn2QQbqA=",
        version = "v0.0.0-20190806162312-597adff16ade",
    )

    go_repository(
        name = "org_golang_x_net",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "golang.org/x/net",
        sum = "h1:pNX+40auqi2JqRfOP1akLGtYcn15TUbkhwuCO3foqqM=",
        version = "v0.0.0-20200602114024-627f9648deb9",
    )

    go_repository(
        name = "org_golang_x_oauth2",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "golang.org/x/oauth2",
        sum = "h1:TzXSXBo42m9gQenoE3b9BGiEpg5IG2JkU5FkPIawgtw=",
        version = "v0.0.0-20200107190931-bf48bf16ab8d",
    )

    go_repository(
        name = "org_golang_x_sync",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "golang.org/x/sync",
        sum = "h1:vcxGaoTs7kV8m5Np9uUNQin4BrLOthgV7252N8V+FwY=",
        version = "v0.0.0-20190911185100-cd5d95a43a6e",
    )

    go_repository(
        name = "org_golang_x_sys",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "golang.org/x/sys",
        sum = "h1:kCCpuwSAoYJPkNc6x0xT9yTtV4oKtARo4RGBQWOfg9E=",
        version = "v0.0.0-20200828194041-157a740278f4",
    )

    go_repository(
        name = "org_golang_x_text",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "golang.org/x/text",
        sum = "h1:cokOdA+Jmi5PJGXLlLllQSgYigAEfHXJAERHVMaCc2k=",
        version = "v0.3.3",
    )

    go_repository(
        name = "org_golang_x_time",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "golang.org/x/time",
        sum = "h1:/5xXl8Y5W96D+TtHSlonuFqGHIWVuyCkGJLwGh9JJFs=",
        version = "v0.0.0-20191024005414-555d28b269f0",
    )

    go_repository(
        name = "org_golang_x_tools",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "golang.org/x/tools",
        sum = "h1:/e4fNMHdLn7SQSxTrRZTma2xjQW6ELdxcnpqMhpo9X4=",
        version = "v0.0.0-20200702044944-0cc1aa72b347",
    )

    go_repository(
        name = "org_golang_x_xerrors",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "golang.org/x/xerrors",
        sum = "h1:E7g+9GITq07hpfrRu66IVDexMakfv52eLZ2CXBWiKr4=",
        version = "v0.0.0-20191204190536-9bdfabe68543",
    )

    go_repository(
        name = "org_gonum_v1_gonum",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "gonum.org/v1/gonum",
        sum = "h1:OB/uP/Puiu5vS5QMRPrXCDWUPb+kt8f1KW8oQzFejQw=",
        version = "v0.0.0-20190331200053-3d26580ed485",
    )

    go_repository(
        name = "org_gonum_v1_netlib",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "gonum.org/v1/netlib",
        sum = "h1:jRyg0XfpwWlhEV8mDfdNGBeSJM2fuyh9Yjrnd8kF2Ts=",
        version = "v0.0.0-20190331212654-76723241ea4e",
    )

    go_repository(
        name = "org_modernc_cc",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "modernc.org/cc",
        sum = "h1:nPibNuDEx6tvYrUAtvDTTw98rx5juGsa5zuDnKwEEQQ=",
        version = "v1.0.0",
    )

    go_repository(
        name = "org_modernc_golex",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "modernc.org/golex",
        sum = "h1:wWpDlbK8ejRfSyi0frMyhilD3JBvtcx2AdGDnU+JtsE=",
        version = "v1.0.0",
    )

    go_repository(
        name = "org_modernc_mathutil",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "modernc.org/mathutil",
        sum = "h1:93vKjrJopTPrtTNpZ8XIovER7iCIH1QU7wNbOQXC60I=",
        version = "v1.0.0",
    )

    go_repository(
        name = "org_modernc_strutil",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "modernc.org/strutil",
        sum = "h1:XVFtQwFVwc02Wk+0L/Z/zDDXO81r5Lhe6iMKmGX3KhE=",
        version = "v1.0.0",
    )

    go_repository(
        name = "org_modernc_xc",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "modernc.org/xc",
        sum = "h1:7ccXrupWZIS3twbUGrtKmHS2DXY6xegFua+6O3xgAFU=",
        version = "v1.0.0",
    )

    go_repository(
        name = "org_mongodb_go_mongo_driver",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "go.mongodb.org/mongo-driver",
        sum = "h1:jxcFYjlkl8xaERsgLo+RNquI0epW6zuy/ZRQs6jnrFA=",
        version = "v1.1.2",
    )

    go_repository(
        name = "org_uber_go_atomic",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "go.uber.org/atomic",
        sum = "h1:Ezj3JGmsOnG1MoRWQkPBsKLe9DwWD9QeXzTRzzldNVk=",
        version = "v1.6.0",
    )

    go_repository(
        name = "org_uber_go_multierr",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "go.uber.org/multierr",
        sum = "h1:KCa4XfM8CWFCpxXRGok+Q0SS/0XBhMDbHHGABQLvD2A=",
        version = "v1.5.0",
    )

    go_repository(
        name = "org_uber_go_zap",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "go.uber.org/zap",
        sum = "h1:nYDKopTbvAPq/NrUVZwT15y2lpROBiLLyoRTbXOYWOo=",
        version = "v1.14.1",
    )

    go_repository(
        name = "tools_gotest",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "gotest.tools",
        sum = "h1:VsBPFP1AI068pPrMxtb/S8Zkgf9xEmTLJjfM+P5UIEo=",
        version = "v2.2.0+incompatible",
    )

    go_repository(
        name = "xyz_gomodules_jsonpatch_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "gomodules.xyz/jsonpatch/v2",
        sum = "h1:Phva6wqu+xR//Njw6iorylFFgn/z547tw5Ne3HZPQ+k=",
        version = "v2.1.0",
    )
    go_repository(
        name = "ag_pack_amqp",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "pack.ag/amqp",
        sum = "h1:cuNDWLUTbKRtEZwhB0WQBXf9pGbm87pUBXQhvcFxBWg=",
        version = "v0.11.2",
    )
    go_repository(
        name = "cc_mvdan_xurls_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "mvdan.cc/xurls/v2",
        sum = "h1:r1zSOSNS/kqtpmATyMMMvaZ4/djsesbYz5kr0+qMRWc=",
        version = "v2.0.0",
    )
    go_repository(
        name = "com_github_alcortesm_tgz",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/alcortesm/tgz",
        sum = "h1:uSoVVbwJiQipAclBbw+8quDsfcvFjOpI5iCf4p/cqCs=",
        version = "v0.0.0-20161220082320-9c5fe88206d7",
    )
    go_repository(
        name = "com_github_andybalholm_brotli",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/andybalholm/brotli",
        sum = "h1:bZ28Hqta7TFAK3Q08CMvv8y3/8ATaEqv2nGoc6yff6c=",
        version = "v0.0.0-20190621154722-5f990b63d2d6",
    )
    go_repository(
        name = "com_github_andygrunwald_go_gerrit",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/andygrunwald/go-gerrit",
        sum = "h1:uUuUZipfD5nPl2L/i0I3N4iRKJcoO2CPjktaH/kP9gQ=",
        version = "v0.0.0-20190120104749-174420ebee6c",
    )
    go_repository(
        name = "com_github_anmitsu_go_shlex",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/anmitsu/go-shlex",
        sum = "h1:kFOfPq6dUM1hTo4JG6LR5AXSUEsOjtdm0kw0FtQtMJA=",
        version = "v0.0.0-20161002113705-648efa622239",
    )
    go_repository(
        name = "com_github_apache_thrift",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/apache/thrift",
        sum = "h1:pODnxUFNcjP9UTLZGTdeh+j16A8lJbRvD3rOtrk/7bs=",
        version = "v0.12.0",
    )
    go_repository(
        name = "com_github_armon_go_socks5",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/armon/go-socks5",
        sum = "h1:0CwZNZbxp69SHPdPJAN/hZIm0C4OItdklCFmMRWYpio=",
        version = "v0.0.0-20160902184237-e75332964ef5",
    )
    go_repository(
        name = "com_github_aws_aws_k8s_tester",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/aws/aws-k8s-tester",
        sum = "h1:Zr5NWiRK5fhmRIlhrsTwrY8yB488FyN6iulci2D7VaI=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_aws_aws_sdk_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/aws/aws-sdk-go",
        sum = "h1:2/sOfb9oPHTRZ0lxinoaTPDcYwNa1H/SpKP4nVRBwmg=",
        version = "v1.33.8",
    )
    go_repository(
        name = "com_github_azure_azure_amqp_common_go_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/Azure/azure-amqp-common-go/v2",
        sum = "h1:+QbFgmWCnPzdaRMfsI0Yb6GrRdBj5jVL8N3EXuEUcBQ=",
        version = "v2.1.0",
    )
    go_repository(
        name = "com_github_azure_azure_pipeline_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/Azure/azure-pipeline-go",
        sum = "h1:7U9HBg1JFK3jHl5qmo4CTZKFTVgMwdFHMVtCdfBE21U=",
        version = "v0.2.3",
    )
    go_repository(
        name = "com_github_azure_azure_sdk_for_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/Azure/azure-sdk-for-go",
        sum = "h1:3D2O4g8AwDwyWkM1HpMFVux/ccQJmGJHXsE004Wsu1Q=",
        version = "v38.0.0+incompatible",
    )
    go_repository(
        name = "com_github_azure_azure_service_bus_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/Azure/azure-service-bus-go",
        sum = "h1:G1qBLQvHCFDv9pcpgwgFkspzvnGknJRR0PYJ9ytY/JA=",
        version = "v0.9.1",
    )
    go_repository(
        name = "com_github_azure_azure_storage_blob_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/Azure/azure-storage-blob-go",
        sum = "h1:7bFXA1QB+lOK2/ASWHhp6/vnxjaeeZq6t8w1Jyp0Iaw=",
        version = "v0.12.0",
    )
    go_repository(
        name = "com_github_azure_go_autorest",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/Azure/go-autorest",
        sum = "h1:V5VMDjClD3GiElqLWO7mz2MxNAK/vTfRHdAubSIPRgs=",
        version = "v14.2.0+incompatible",
    )
    go_repository(
        name = "com_github_azure_go_autorest_autorest_to",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/Azure/go-autorest/autorest/to",
        sum = "h1:zebkZaadz7+wIQYgC7GXaz3Wb28yKYfVkkBKwc38VF8=",
        version = "v0.3.0",
    )
    go_repository(
        name = "com_github_azure_go_autorest_autorest_validation",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/Azure/go-autorest/autorest/validation",
        sum = "h1:15vMO4y76dehZSq7pAaOLQxC6dZYsSrj2GQpflyM/L4=",
        version = "v0.2.0",
    )
    go_repository(
        name = "com_github_bazelbuild_buildtools",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/bazelbuild/buildtools",
        sum = "h1:3B/ZE1a6eEJ/4Jf/M6RM2KBouN8yKCUcMmXzSyWqa3g=",
        version = "v0.0.0-20190917191645-69366ca98f89",
    )
    go_repository(
        name = "com_github_bshuster_repo_logrus_logstash_hook",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/bshuster-repo/logrus-logstash-hook",
        sum = "h1:pgAtgj+A31JBVtEHu2uHuEx0n+2ukqUJnS2vVe5pQNA=",
        version = "v0.4.1",
    )
    go_repository(
        name = "com_github_bugsnag_osext",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/bugsnag/osext",
        sum = "h1:otBG+dV+YK+Soembjv71DPz3uX/V/6MMlSyD9JBQ6kQ=",
        version = "v0.0.0-20130617224835-0dd3f918b21b",
    )
    go_repository(
        name = "com_github_bwmarrin_snowflake",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/bwmarrin/snowflake",
        sum = "h1:dRbqXFjM10uA3wdrVZ8Kh19uhciRMOroUYJ7qAqDLhY=",
        version = "v0.0.0",
    )
    go_repository(
        name = "com_github_census_instrumentation_opencensus_proto",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/census-instrumentation/opencensus-proto",
        sum = "h1:glEXhBS5PSLLv4IXzLA5yPRVX4bilULVyxxbrfOtDAk=",
        version = "v0.2.1",
    )
    go_repository(
        name = "com_github_cespare_xxhash",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/cespare/xxhash",
        sum = "h1:a6HrQnmkObjyL+Gs60czilIUGqrzKutQD6XZog3p+ko=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_chai2010_gettext_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/chai2010/gettext-go",
        sum = "h1:7aWHqerlJ41y6FOsEUvknqgXnGmJyJSbjhAWq5pO4F8=",
        version = "v0.0.0-20160711120539-c6fed771bfd5",
    )
    go_repository(
        name = "com_github_cihub_seelog",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/cihub/seelog",
        sum = "h1:kHaBemcxl8o/pQ5VM1c8PVE1PubbNx3mjUr09OqWGCs=",
        version = "v0.0.0-20170130134532-f561c5e57575",
    )
    go_repository(
        name = "com_github_clarketm_json",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/clarketm/json",
        sum = "h1:0JketcMdLC16WGnRGJiNmTXuQznDEQaiknxSPRBxg+k=",
        version = "v1.13.4",
    )
    go_repository(
        name = "com_github_cloudevents_sdk_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/cloudevents/sdk-go",
        sum = "h1:gS5I0s2qPmdc4GBPlUmzZU7RH30BaiOdcRJ1RkXnPrc=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_containerd_cgroups",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/containerd/cgroups",
        sum = "h1:tSNMc+rJDfmYntojat8lljbt1mgKNpTxUZJsSzJ9Y1s=",
        version = "v0.0.0-20190919134610-bf292b21730f",
    )
    go_repository(
        name = "com_github_containerd_console",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/containerd/console",
        sum = "h1:uict5mhHFTzKLUCufdSLym7z/J0CbBJT59lYbP9wtbg=",
        version = "v0.0.0-20180822173158-c12b1e7919c1",
    )
    go_repository(
        name = "com_github_containerd_containerd",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/containerd/containerd",
        sum = "h1:LoIzb5y9x5l8VKAlyrbusNPXqBY0+kviRloxFUMFwKc=",
        version = "v1.3.3",
    )
    go_repository(
        name = "com_github_containerd_continuity",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/containerd/continuity",
        sum = "h1:kIFnQBO7rQ0XkMe6xEwbybYHBEaWmh/f++laI6Emt7M=",
        version = "v0.0.0-20200107194136-26c1120b8d41",
    )
    go_repository(
        name = "com_github_containerd_fifo",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/containerd/fifo",
        sum = "h1:PUD50EuOMkXVcpBIA/R95d56duJR9VxhwncsFbNnxW4=",
        version = "v0.0.0-20190226154929-a9fb20d87448",
    )
    go_repository(
        name = "com_github_containerd_go_runc",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/containerd/go-runc",
        sum = "h1:esQOJREg8nw8aXj6uCN5dfW5cKUBiEJ/+nni1Q/D/sw=",
        version = "v0.0.0-20180907222934-5a6d9f37cfa3",
    )
    go_repository(
        name = "com_github_containerd_ttrpc",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/containerd/ttrpc",
        sum = "h1:dlfGmNcE3jDAecLqwKPMNX6nk2qh1c1Vg1/YTzpOOF4=",
        version = "v0.0.0-20190828154514-0e0f228740de",
    )
    go_repository(
        name = "com_github_containerd_typeurl",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/containerd/typeurl",
        sum = "h1:JNn81o/xG+8NEo3bC/vx9pbi/g2WI8mtP2/nXzu297Y=",
        version = "v0.0.0-20180627222232-a93fcdb778cd",
    )
    go_repository(
        name = "com_github_coreos_bbolt",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/coreos/bbolt",
        sum = "h1:wZwiHHUieZCquLkDL0B8UhzreNWsPHooDAG3q34zk0s=",
        version = "v1.3.2",
    )
    go_repository(
        name = "com_github_cpuguy83_go_md2man_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/cpuguy83/go-md2man/v2",
        sum = "h1:EoUDS0afbrsXAZ9YQ9jdu/mZ2sXgT1/2yyNng4PGlyM=",
        version = "v2.0.0",
    )
    go_repository(
        name = "com_github_cyphar_filepath_securejoin",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/cyphar/filepath-securejoin",
        sum = "h1:jCwT2GTP+PY5nBz3c/YL5PAIbusElVrPujOBSCj8xRg=",
        version = "v0.2.2",
    )
    go_repository(
        name = "com_github_datadog_zstd",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/DataDog/zstd",
        sum = "h1:+IawcoXhCBylN7ccwdwf8LOH2jKq7NavGpEPanrlTzE=",
        version = "v1.4.4",
    )
    go_repository(
        name = "com_github_daviddengcn_go_colortext",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/daviddengcn/go-colortext",
        sum = "h1:uVsMphB1eRx7xB1njzL3fuMdWRN8HtVzoUOItHMwv5c=",
        version = "v0.0.0-20160507010035-511bcaf42ccd",
    )
    go_repository(
        name = "com_github_deislabs_oras",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/deislabs/oras",
        sum = "h1:If674KraJVpujYR00rzdi0QAmW4BxzMJPVAZJKuhQ0c=",
        version = "v0.8.1",
    )
    go_repository(
        name = "com_github_denisenkom_go_mssqldb",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/denisenkom/go-mssqldb",
        sum = "h1:83Wprp6ROGeiHFAP8WJdI2RoxALQYgdllERc3N5N2DM=",
        version = "v0.0.0-20191124224453-732737034ffd",
    )
    go_repository(
        name = "com_github_denverdino_aliyungo",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/denverdino/aliyungo",
        sum = "h1:p6poVbjHDkKa+wtC8frBMwQtT3BmqGYBjzMwJ63tuR4=",
        version = "v0.0.0-20190125010748-a747050bb1ba",
    )
    go_repository(
        name = "com_github_devigned_tab",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/devigned/tab",
        sum = "h1:3mD6Kb1mUOYeLpJvTVSDwSg5ZsfSxfvxGRTxRsJsITA=",
        version = "v0.1.1",
    )
    go_repository(
        name = "com_github_dgryski_go_sip13",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/dgryski/go-sip13",
        sum = "h1:RMLoZVzv4GliuWafOuPuQDKSm1SJph7uCRnnS61JAn4=",
        version = "v0.0.0-20181026042036-e10d5fee7954",
    )
    go_repository(
        name = "com_github_dimchansky_utfbom",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/dimchansky/utfbom",
        sum = "h1:FcM3g+nofKgUteL8dm/UpdRXNC9KmADgTpLKsu0TRo4=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_djherbis_atime",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/djherbis/atime",
        sum = "h1:ySLvBAM0EvOGaX7TI4dAM5lWj+RdJUCKtGSEHN8SGBg=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_dnaeon_go_vcr",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/dnaeon/go-vcr",
        sum = "h1:r8L/HqC0Hje5AXMu1ooW8oyQyOFv4GxqpL0nRP7SLLY=",
        version = "v1.0.1",
    )
    go_repository(
        name = "com_github_docker_cli",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/docker/cli",
        sum = "h1:FwssHbCDJD025h+BchanCwE1Q8fyMgqDr2mOQAWOLGw=",
        version = "v0.0.0-20200130152716-5d0cf8839492",
    )
    go_repository(
        name = "com_github_docker_distribution",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/docker/distribution",
        sum = "h1:a5mlkVzth6W5A4fOsS3D2EO5BUmsJpcB+cRlLU7cSug=",
        version = "v2.7.1+incompatible",
    )
    go_repository(
        name = "com_github_docker_docker_credential_helpers",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/docker/docker-credential-helpers",
        sum = "h1:zI2p9+1NQYdnG6sMU26EX4aVGlqbInSQxQXLvzJ4RPQ=",
        version = "v0.6.3",
    )
    go_repository(
        name = "com_github_docker_go_connections",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/docker/go-connections",
        sum = "h1:El9xVISelRB7BuFusrZozjnkIM5YnzCViNKohAFqRJQ=",
        version = "v0.4.0",
    )
    go_repository(
        name = "com_github_docker_go_metrics",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/docker/go-metrics",
        sum = "h1:yWHOI+vFjEsAakUTSrtqc/SAHrhSkmn48pqjidZX3QA=",
        version = "v0.0.0-20180209012529-399ea8c73916",
    )
    go_repository(
        name = "com_github_docker_libtrust",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/docker/libtrust",
        sum = "h1:ZClxb8laGDf5arXfYcAtECDFgAgHklGI8CxgjHnXKJ4=",
        version = "v0.0.0-20150114040149-fa567046d9b1",
    )
    go_repository(
        name = "com_github_dsnet_compress",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/dsnet/compress",
        sum = "h1:PlZu0n3Tuv04TzpfPbrnI0HW/YwodEXDS+oPKahKF0Q=",
        version = "v0.0.1",
    )
    go_repository(
        name = "com_github_dsnet_golib",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/dsnet/golib",
        sum = "h1:tFh1tRc4CA31yP6qDcu+Trax5wW5GuMxvkIba07qVLY=",
        version = "v0.0.0-20171103203638-1ea166775780",
    )
    go_repository(
        name = "com_github_eapache_go_resiliency",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/eapache/go-resiliency",
        sum = "h1:v7g92e/KSN71Rq7vSThKaWIq68fL4YHvWyiUKorFR1Q=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_eapache_go_xerial_snappy",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/eapache/go-xerial-snappy",
        sum = "h1:YEetp8/yCZMuEPMUDHG0CW/brkkEp8mzqk2+ODEitlw=",
        version = "v0.0.0-20180814174437-776d5712da21",
    )
    go_repository(
        name = "com_github_eapache_queue",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/eapache/queue",
        sum = "h1:YOEu7KNc61ntiQlcEeUIoDTJ2o8mQznoNvUhiigpIqc=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_emirpasic_gods",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/emirpasic/gods",
        sum = "h1:QAUIPSaCu4G+POclxeqb3F+WPpdKqFGlw36+yOzGlrg=",
        version = "v1.12.0",
    )
    go_repository(
        name = "com_github_envoyproxy_go_control_plane",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/envoyproxy/go-control-plane",
        sum = "h1:rEvIZUSZ3fx39WIi3JkQqQBitGwpELBIYWeBVh6wn+E=",
        version = "v0.9.4",
    )
    go_repository(
        name = "com_github_envoyproxy_protoc_gen_validate",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/envoyproxy/protoc-gen-validate",
        sum = "h1:EQciDnbrYxy13PgWoY8AqoxGiPrpgBZ1R8UNe3ddc+A=",
        version = "v0.1.0",
    )
    go_repository(
        name = "com_github_erikstmartin_go_testdb",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/erikstmartin/go-testdb",
        sum = "h1:Yzb9+7DPaBjB8zlTR87/ElzFsnQfuHnVUVqpZZIcV5Y=",
        version = "v0.0.0-20160219214506-8d10e4a1bae5",
    )
    go_repository(
        name = "com_github_exponent_io_jsonpath",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/exponent-io/jsonpath",
        sum = "h1:105gxyaGwCFad8crR9dcMQWvV9Hvulu6hwUh4tWPJnM=",
        version = "v0.0.0-20151013193312-d6023ce2651d",
    )
    go_repository(
        name = "com_github_fatih_camelcase",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/fatih/camelcase",
        sum = "h1:hxNvNX/xYBp0ovncs8WyWZrOrpBNub/JfaMvbURyft8=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_flynn_go_shlex",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/flynn/go-shlex",
        sum = "h1:BHsljHzVlRcyQhjrss6TZTdY2VfCqZPbv5k3iBFa2ZQ=",
        version = "v0.0.0-20150515145356-3f9db97f8568",
    )
    go_repository(
        name = "com_github_fortytw2_leaktest",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/fortytw2/leaktest",
        sum = "h1:u8491cBMTQ8ft8aeV+adlcytMZylmA5nnwwkRZjI8vw=",
        version = "v1.3.0",
    )
    go_repository(
        name = "com_github_frankban_quicktest",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/frankban/quicktest",
        sum = "h1:PvpJR0Uq8SdX+zagCMsarBMlhz6ysGTf1+pRmCsRXqY=",
        version = "v1.8.1",
    )
    go_repository(
        name = "com_github_fsouza_fake_gcs_server",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/fsouza/fake-gcs-server",
        sum = "h1:3iml5UHzQtk3cpnYfqW16Ia+1xSuu9tc4BElZu5470M=",
        version = "v0.0.0-20180612165233-e85be23bdaa8",
    )
    go_repository(
        name = "com_github_garyburd_redigo",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/garyburd/redigo",
        sum = "h1:LofdAjjjqCSXMwLGgOgnE+rdPuvX9DxCqaHwKy7i/ko=",
        version = "v0.0.0-20150301180006-535138d7bcd7",
    )
    go_repository(
        name = "com_github_gliderlabs_ssh",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/gliderlabs/ssh",
        sum = "h1:6zsha5zo/TWhRhwqCD3+EarCAgZ2yN28ipRnGPnwkI0=",
        version = "v0.2.2",
    )
    go_repository(
        name = "com_github_go_bindata_go_bindata_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/go-bindata/go-bindata/v3",
        sum = "h1:F0nVttLC3ws0ojc7p60veTurcOm//D4QBODNM7EGrCI=",
        version = "v3.1.3",
    )
    go_repository(
        name = "com_github_go_gl_glfw",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/go-gl/glfw",
        sum = "h1:QbL/5oDUmRBzO9/Z7Seo6zf912W/a6Sr4Eu0G/3Jho0=",
        version = "v0.0.0-20190409004039-e6da0acd62b1",
    )
    go_repository(
        name = "com_github_go_ini_ini",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/go-ini/ini",
        sum = "h1:0wVcG9udk2C3TGgmdIGKK9ScOZHZB5nbG+gwji9fhhc=",
        version = "v1.55.0",
    )
    go_repository(
        name = "com_github_go_sql_driver_mysql",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/go-sql-driver/mysql",
        sum = "h1:ozyZYNQW3x3HtqT1jira07DN2PArx2v7/mN66gGcHOs=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_github_go_test_deep",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/go-test/deep",
        sum = "h1:u2CU3YKy9I2pmu9pX0eq50wCgjfGIt539SqR7FbHiho=",
        version = "v1.0.4",
    )
    go_repository(
        name = "com_github_gobuffalo_envy",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/gobuffalo/envy",
        sum = "h1:X3is06x7v0nW2xiy2yFbbIjwHz57CD6z6MkvqULTCm8=",
        version = "v1.6.5",
    )
    go_repository(
        name = "com_github_gobwas_glob",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/gobwas/glob",
        sum = "h1:A4xDbljILXROh+kObIiy5kIaPYD8e96x1tgBhUI5J+Y=",
        version = "v0.2.3",
    )
    go_repository(
        name = "com_github_godbus_dbus",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/godbus/dbus",
        sum = "h1:BWhy2j3IXJhjCbC68FptL43tDKIq8FladmaTs3Xs7Z8=",
        version = "v0.0.0-20190422162347-ade71ed3457e",
    )
    go_repository(
        name = "com_github_gofrs_flock",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/gofrs/flock",
        sum = "h1:DP+LD/t0njgoPBvT5MJLeliUIVQR03hiKR6vezdwHlc=",
        version = "v0.7.1",
    )
    go_repository(
        name = "com_github_gogo_googleapis",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/gogo/googleapis",
        sum = "h1:kFkMAZBNAn4j7K0GiZr8cRYzejq68VbheufiV3YuyFI=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_golang_gddo",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/golang/gddo",
        sum = "h1:KRMr9A3qfbVM7iV/WcLY/rL5LICqwMHLhwRXKu99fXw=",
        version = "v0.0.0-20190419222130-af0f2af80721",
    )
    go_repository(
        name = "com_github_golang_snappy",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/golang/snappy",
        sum = "h1:gFVkHXmVAhEbxZVDln5V9GKrLaluNoFHDbrZwAWZgws=",
        version = "v0.0.2-0.20190904063534-ff6b7dc882cf",
    )
    go_repository(
        name = "com_github_golang_sql_civil",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/golang-sql/civil",
        sum = "h1:lXe2qZdvpiX5WZkZR4hgp4KJVfY3nMkvmwbVkpv1rVY=",
        version = "v0.0.0-20190719163853-cb61b32ac6fe",
    )
    go_repository(
        name = "com_github_golangplus_bytes",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/golangplus/bytes",
        sum = "h1:7xqw01UYS+KCI25bMrPxwNYkSns2Db1ziQPpVq99FpE=",
        version = "v0.0.0-20160111154220-45c989fe5450",
    )
    go_repository(
        name = "com_github_golangplus_fmt",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/golangplus/fmt",
        sum = "h1:f5gsjBiF9tRRVomCvrkGMMWI8W1f2OBFar2c5oakAP0=",
        version = "v0.0.0-20150411045040-2a5d6d7d2995",
    )
    go_repository(
        name = "com_github_golangplus_testing",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/golangplus/testing",
        sum = "h1:KhcknUwkWHKZPbFy2P7jH5LKJ3La+0ZeknkkmrSgqb0=",
        version = "v0.0.0-20180327235837-af21d9c3145e",
    )
    go_repository(
        name = "com_github_gomodule_redigo",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/gomodule/redigo",
        sum = "h1:y0Wmhvml7cGnzPa9nocn/fMraMH/lMDdeG+rkx4VgYY=",
        version = "v1.7.1-0.20190724094224-574c33c3df38",
    )
    go_repository(
        name = "com_github_google_go_containerregistry",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/google/go-containerregistry",
        sum = "h1:oGqapkPUiypdS9ch/Vu0npPe03RQ0BhVDYli+OEKNAA=",
        version = "v0.0.0-20200115214256-379933c9c22b",
    )
    go_repository(
        name = "com_github_google_go_github",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/google/go-github",
        sum = "h1:N0LgJ1j65A7kfXrZnUDaYCs/Sf4rEjNlfyDHW9dolSY=",
        version = "v17.0.0+incompatible",
    )
    go_repository(
        name = "com_github_google_go_licenses",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/google/go-licenses",
        sum = "h1:ZIb3nb+/mHAGRkyuxfPykmYdUi21mr8YTGpr/xGPJ8o=",
        version = "v0.0.0-20191112164736-212ea350c932",
    )
    go_repository(
        name = "com_github_google_go_querystring",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/google/go-querystring",
        sum = "h1:Xkwi/a1rcvNg1PPYe5vI8GbeBY/jrVuDX5ASuANWTrk=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_google_go_replayers_grpcreplay",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/google/go-replayers/grpcreplay",
        sum = "h1:eNb1y9rZFmY4ax45uEEECSa8fsxGRU+8Bil52ASAwic=",
        version = "v0.1.0",
    )
    go_repository(
        name = "com_github_google_go_replayers_httpreplay",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/google/go-replayers/httpreplay",
        sum = "h1:AX7FUb4BjrrzNvblr/OlgwrmFiep6soj5K2QSDW7BGk=",
        version = "v0.1.0",
    )
    go_repository(
        name = "com_github_google_licenseclassifier",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/google/licenseclassifier",
        sum = "h1:nVgx26pAe6l/02mYomOuZssv28XkacGw/0WeiTVorqw=",
        version = "v0.0.0-20190926221455-842c0d70d702",
    )
    go_repository(
        name = "com_github_google_renameio",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/google/renameio",
        sum = "h1:GOZbcHa3HfsPKPlmyPyN2KEohoMXOhdMbHrvbpl2QaA=",
        version = "v0.1.0",
    )
    go_repository(
        name = "com_github_google_subcommands",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/google/subcommands",
        sum = "h1:/eqq+otEXm5vhfBrbREPCSVQbvofip6kIz+mX5TUH7k=",
        version = "v1.0.1",
    )
    go_repository(
        name = "com_github_google_wire",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/google/wire",
        sum = "h1:imGQZGEVEHpje5056+K+cgdO72p0LQv2xIIFXNGUf60=",
        version = "v0.3.0",
    )
    go_repository(
        name = "com_github_googleapis_gax_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/googleapis/gax-go",
        sum = "h1:silFMLAnr330+NRuag/VjIGF7TLp/LBrV2CJKFLWEww=",
        version = "v2.0.2+incompatible",
    )
    go_repository(
        name = "com_github_googlecloudplatform_cloud_builders_gcs_fetcher",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/GoogleCloudPlatform/cloud-builders/gcs-fetcher",
        sum = "h1:Pjo3SOZigEnIGevhFqcbFndnqyCH8WimcREd3hRM9vU=",
        version = "v0.0.0-20191203181535-308b93ad1f39",
    )
    go_repository(
        name = "com_github_googlecloudplatform_cloudsql_proxy",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/GoogleCloudPlatform/cloudsql-proxy",
        sum = "h1:sTOp2Ajiew5XIH92YSdwhYc+bgpUX5j5TKK/Ac8Saw8=",
        version = "v0.0.0-20191009163259-e802c2cb94ae",
    )
    go_repository(
        name = "com_github_googlecloudplatform_k8s_cloud_provider",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/GoogleCloudPlatform/k8s-cloud-provider",
        sum = "h1:N7lSsF+R7wSulUADi36SInSQA3RvfO/XclHQfedr0qk=",
        version = "v0.0.0-20190822182118-27a4ced34534",
    )
    go_repository(
        name = "com_github_googlecloudplatform_testgrid",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/GoogleCloudPlatform/testgrid",
        sum = "h1:tKM75ScxinVqDkguwG5AnsQZn9XCNpxAdEey1OfJgiE=",
        version = "v0.0.7",
    )
    go_repository(
        name = "com_github_gopherjs_gopherjs",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/gopherjs/gopherjs",
        sum = "h1:EGx4pi6eqNxGaHF6qqu48+N2wcFQ5qg5FXgOdqsJ5d8=",
        version = "v0.0.0-20181017120253-0766667cb4d1",
    )
    go_repository(
        name = "com_github_gorilla_context",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/gorilla/context",
        sum = "h1:AWwleXJkX/nhcU9bZSnZoi3h/qGYqQAGhq6zZe/aQW8=",
        version = "v1.1.1",
    )
    go_repository(
        name = "com_github_gorilla_csrf",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/gorilla/csrf",
        sum = "h1:QqQ/OWwuFp4jMKgBFAzJVW3FMULdyUW7JoM4pEWuqKg=",
        version = "v1.6.2",
    )
    go_repository(
        name = "com_github_gorilla_handlers",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/gorilla/handlers",
        sum = "h1:893HsJqtxp9z1SF76gg6hY70hRY1wVlTSnC/h1yUDCo=",
        version = "v0.0.0-20150720190736-60c7bfde3e33",
    )
    go_repository(
        name = "com_github_gorilla_mux",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/gorilla/mux",
        sum = "h1:VuZ8uybHlWmqV03+zRzdwKL4tUnIp1MAQtp1mIFE1bc=",
        version = "v1.7.4",
    )
    go_repository(
        name = "com_github_gorilla_securecookie",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/gorilla/securecookie",
        sum = "h1:miw7JPhV+b/lAHSXz4qd/nN9jRiAFV5FwjeKyCS8BvQ=",
        version = "v1.1.1",
    )
    go_repository(
        name = "com_github_gorilla_sessions",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/gorilla/sessions",
        sum = "h1:S7P+1Hm5V/AT9cjEcUD5uDaQSX0OE577aCXgoaKpYbQ=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_gosuri_uitable",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/gosuri/uitable",
        sum = "h1:IG2xLKRvErL3uhY6e1BylFzG+aJiwQviDDTfOKeKTpY=",
        version = "v0.0.4",
    )
    go_repository(
        name = "com_github_h2non_gock",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/h2non/gock",
        sum = "h1:17gCehSo8ZOgEsFKpQgqHiR7VLyjxdAG3lkhVvO9QZU=",
        version = "v1.0.9",
    )
    go_repository(
        name = "com_github_hashicorp_errwrap",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/hashicorp/errwrap",
        sum = "h1:hLrqtEDnRye3+sgx6z4qVLNuviH3MR5aQ0ykNJa/UYA=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_hashicorp_go_multierror",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/hashicorp/go-multierror",
        sum = "h1:iVjPR7a6H0tWELX5NxNe7bYopibicUzc7uPribsnS6o=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_hashicorp_go_uuid",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/hashicorp/go-uuid",
        sum = "h1:cfejS+Tpcp13yd5nYHWDI6qVCny6wyX2Mt5SGur2IGE=",
        version = "v1.0.2",
    )
    go_repository(
        name = "com_github_huandu_xstrings",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/huandu/xstrings",
        sum = "h1:gvV6jG9dTgFEncxo+AF7PH6MZXi/vZl25owA/8Dg8Wo=",
        version = "v1.3.0",
    )
    go_repository(
        name = "com_github_influxdata_influxdb",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/influxdata/influxdb",
        sum = "h1:AciJ2ei/llFRundm7CtqwF6B2aOds1A7QG3sMW8QiaQ=",
        version = "v0.0.0-20161215172503-049f9b42e9a5",
    )
    go_repository(
        name = "com_github_jbenet_go_context",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/jbenet/go-context",
        sum = "h1:BQSFePA1RWJOlocH6Fxy8MmwDt+yVQYULKfN0RoTN8A=",
        version = "v0.0.0-20150711004518-d14ea06fba99",
    )
    go_repository(
        name = "com_github_jcmturner_gofork",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/jcmturner/gofork",
        sum = "h1:J7uCkflzTEhUZ64xqKnkDxq3kzc96ajM1Gli5ktUem8=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_jenkins_x_go_scm",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/jenkins-x/go-scm",
        sum = "h1:+HIEkc/Dzdq0buJF8q0Keip2wexW40BfkrDXKx88T78=",
        version = "v1.5.79",
    )
    go_repository(
        name = "com_github_jessevdk_go_flags",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/jessevdk/go-flags",
        sum = "h1:4IU2WS7AumrZ/40jfhf4QVDMsQwqA7VEHozFRrGARJA=",
        version = "v1.4.0",
    )
    go_repository(
        name = "com_github_jinzhu_gorm",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/jinzhu/gorm",
        sum = "h1:Drgk1clyWT9t9ERbzHza6Mj/8FY/CqMyVzOiHviMo6Q=",
        version = "v1.9.12",
    )
    go_repository(
        name = "com_github_jinzhu_inflection",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/jinzhu/inflection",
        sum = "h1:K317FqzuhWc8YvSVlFMCCUb36O/S9MCKRDI7QkRKD/E=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_jinzhu_now",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/jinzhu/now",
        sum = "h1:g39TucaRWyV3dwDO++eEc6qf8TVIQ/Da48WmqjZ3i7E=",
        version = "v1.1.1",
    )
    go_repository(
        name = "com_github_jmespath_go_jmespath",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/jmespath/go-jmespath",
        sum = "h1:OS12ieG61fsCg5+qLJ+SsW9NicxNkg3b25OyT2yCeUc=",
        version = "v0.3.0",
    )
    go_repository(
        name = "com_github_joefitzgerald_rainbow_reporter",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/joefitzgerald/rainbow-reporter",
        sum = "h1:AuMG652zjdzI0YCCnXAqATtRBpGXMcAnrajcaTrSeuo=",
        version = "v0.1.0",
    )
    go_repository(
        name = "com_github_joho_godotenv",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/joho/godotenv",
        sum = "h1:Zjp+RcGpHhGlrMbJzXTrZZPrWj+1vfm90La1wgB6Bhc=",
        version = "v1.3.0",
    )
    go_repository(
        name = "com_github_jtolds_gls",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/jtolds/gls",
        sum = "h1:xdiiI2gbIgH/gLH7ADydsJ1uDOEzR8yvV7C0MuV77Wo=",
        version = "v4.20.0+incompatible",
    )
    go_repository(
        name = "com_github_kelseyhightower_envconfig",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/kelseyhightower/envconfig",
        sum = "h1:Im6hONhd3pLkfDFsbRgu68RDNkGF1r3dvMUtDTo2cv8=",
        version = "v1.4.0",
    )
    go_repository(
        name = "com_github_kevinburke_ssh_config",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/kevinburke/ssh_config",
        sum = "h1:Coekwdh0v2wtGp9Gmz1Ze3eVRAWJMLokvN3QjdzCHLY=",
        version = "v0.0.0-20190725054713-01f96b0aa0cd",
    )
    go_repository(
        name = "com_github_klauspost_compress",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/klauspost/compress",
        sum = "h1:Znfn6hXZAHaLPNnlqUYRrBSReFHYybslgv4PTiyz6P0=",
        version = "v1.10.2",
    )
    go_repository(
        name = "com_github_klauspost_cpuid",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/klauspost/cpuid",
        sum = "h1:vJi+O/nMdFt0vqm8NZBI6wzALWdA2X+egi0ogNyrC/w=",
        version = "v1.2.1",
    )
    go_repository(
        name = "com_github_klauspost_pgzip",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/klauspost/pgzip",
        sum = "h1:oIPZROsWuPHpOdMVWLuJZXwgjhrW8r1yEX8UqMyeNHM=",
        version = "v1.2.1",
    )
    go_repository(
        name = "com_github_lib_pq",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/lib/pq",
        sum = "h1:9xohqzkUwzR4Ga4ivdTcawVS89YSDVxXMa3xJX3cGzg=",
        version = "v1.8.0",
    )
    go_repository(
        name = "com_github_liggitt_tabwriter",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/liggitt/tabwriter",
        sum = "h1:9TO3cAIGXtEhnIaL+V+BEER86oLrvS+kWobKpbJuye0=",
        version = "v0.0.0-20181228230101-89fcab3d43de",
    )
    go_repository(
        name = "com_github_lithammer_dedent",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/lithammer/dedent",
        sum = "h1:VNzHMVCBNG1j0fh3OrsFRkVUwStdDArbgBWoPAffktY=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_lyft_protoc_gen_validate",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/lyft/protoc-gen-validate",
        sum = "h1:KNt/RhmQTOLr7Aj8PsJ7mTronaFyx80mRTT9qF261dA=",
        version = "v0.0.13",
    )
    go_repository(
        name = "com_github_makenowjust_heredoc",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/MakeNowJust/heredoc",
        sum = "h1:sjQovDkwrZp8u+gxLtPgKGjk5hCxuy2hrRejBTA9xFU=",
        version = "v0.0.0-20170808103936-bb23615498cd",
    )
    go_repository(
        name = "com_github_markbates_inflect",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/markbates/inflect",
        sum = "h1:5fh1gzTFhfae06u3hzHYO9xe3l3v3nW5Pwt3naLTP5g=",
        version = "v1.0.4",
    )
    go_repository(
        name = "com_github_marstr_guid",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/marstr/guid",
        sum = "h1:/M4H/1G4avsieL6BbUwCOBzulmoeKVP5ux/3mQNnbyI=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_masterminds_goutils",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/Masterminds/goutils",
        sum = "h1:zukEsf/1JZwCMgHiK3GZftabmxiCw4apj3a28RPBiVg=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_masterminds_sprig_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/Masterminds/sprig/v3",
        sum = "h1:wz22D0CiSctrliXiI9ZO3HoNApweeRGftyDN+BQa3B8=",
        version = "v3.0.2",
    )
    go_repository(
        name = "com_github_masterminds_vcs",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/Masterminds/vcs",
        sum = "h1:NL3G1X7/7xduQtA2sJLpVpfHTNBALVNSjob6KEjPXNQ=",
        version = "v1.13.1",
    )
    go_repository(
        name = "com_github_mattbaird_jsonpatch",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/mattbaird/jsonpatch",
        sum = "h1:+J2gw7Bw77w/fbK7wnNJJDKmw1IbWft2Ul5BzrG1Qm8=",
        version = "v0.0.0-20171005235357-81af80346b1a",
    )
    go_repository(
        name = "com_github_mattn_go_ieproxy",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/mattn/go-ieproxy",
        sum = "h1:qiyop7gCflfhwCzGyeT0gro3sF9AIg9HU98JORTkqfI=",
        version = "v0.0.1",
    )
    go_repository(
        name = "com_github_mattn_go_shellwords",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/mattn/go-shellwords",
        sum = "h1:eaB5JspOwiKKcHdqcjbfe5lA9cNn/4NRRtddXJCimqk=",
        version = "v1.0.9",
    )
    go_repository(
        name = "com_github_mattn_go_sqlite3",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/mattn/go-sqlite3",
        sum = "h1:xQ15muvnzGBHpIpdrNi1DA5x0+TcBZzsIDwmw9uTHzw=",
        version = "v2.0.1+incompatible",
    )
    go_repository(
        name = "com_github_mattn_go_zglob",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/mattn/go-zglob",
        sum = "h1:6HgbBMgs3hI9y1/MYG0r9j6daUubUskZNsEW4fkWR/k=",
        version = "v0.0.2-0.20191112051448-a8912a37f9e7",
    )
    go_repository(
        name = "com_github_maxbrunsfeld_counterfeiter_v6",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/maxbrunsfeld/counterfeiter/v6",
        sum = "h1:g+4J5sZg6osfvEfkRZxJ1em0VT95/UOZgi/l7zi1/oE=",
        version = "v6.2.2",
    )
    go_repository(
        name = "com_github_mholt_archiver_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/mholt/archiver/v3",
        sum = "h1:vWjhY8SQp5yzM9P6OJ/eZEkmi3UAbRrxCq48MxjAzig=",
        version = "v3.3.0",
    )
    go_repository(
        name = "com_github_microsoft_go_winio",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/Microsoft/go-winio",
        sum = "h1:ygIc8M6trr62pF5DucadTWGdEB4mEyvzi0e2nbcmcyA=",
        version = "v0.4.15-0.20190919025122-fc70bd9a86b5",
    )
    go_repository(
        name = "com_github_microsoft_hcsshim",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/Microsoft/hcsshim",
        sum = "h1:ptnOoufxGSzauVTsdE+wMYnCWA301PdoN4xg5oRdZpg=",
        version = "v0.8.7",
    )
    go_repository(
        name = "com_github_mitchellh_copystructure",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/mitchellh/copystructure",
        sum = "h1:Laisrj+bAB6b/yJwB5Bt3ITZhGJdqmxquMKeZ+mmkFQ=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_mitchellh_go_wordwrap",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/mitchellh/go-wordwrap",
        sum = "h1:6GlHJ/LTGMrIJbwgdqdl2eEH8o+Exx/0m8ir9Gns0u4=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_mitchellh_ioprogress",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/mitchellh/ioprogress",
        sum = "h1:Qa6dnn8DlasdXRnacluu8HzPts0S1I9zvvUPDbBnXFI=",
        version = "v0.0.0-20180201004757-6a23b12fa88e",
    )
    go_repository(
        name = "com_github_mitchellh_osext",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/mitchellh/osext",
        sum = "h1:2+myh5ml7lgEU/51gbeLHfKGNfgEQQIWrlbdaOsidbQ=",
        version = "v0.0.0-20151018003038-5e2d6d41470f",
    )
    go_repository(
        name = "com_github_mitchellh_reflectwalk",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/mitchellh/reflectwalk",
        sum = "h1:9D+8oIskB4VJBN5SFlmc27fSlIBZaov1Wpk/IfikLNY=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_mohae_deepcopy",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/mohae/deepcopy",
        sum = "h1:RWengNIwukTxcDr9M+97sNutRR1RKhG96O6jWumTTnw=",
        version = "v0.0.0-20170929034955-c48cc78d4826",
    )
    go_repository(
        name = "com_github_morikuni_aec",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/morikuni/aec",
        sum = "h1:nP9CBfwrvYnBRgY6qfDQkygYDmYwOilePFkwzv4dU8A=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_nats_io_jwt",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/nats-io/jwt",
        sum = "h1:+RB5hMpXUUA2dfxuhBTEkMOrYmM+gKIZYS1KjSostMI=",
        version = "v0.3.2",
    )
    go_repository(
        name = "com_github_nats_io_nats_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/nats-io/nats.go",
        sum = "h1:ik3HbLhZ0YABLto7iX80pZLPw/6dx3T+++MZJwLnMrQ=",
        version = "v1.9.1",
    )
    go_repository(
        name = "com_github_nats_io_nats_server_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/nats-io/nats-server/v2",
        sum = "h1:i2Ly0B+1+rzNZHHWtD4ZwKi+OU5l+uQo1iDHZ2PmiIc=",
        version = "v2.1.2",
    )
    go_repository(
        name = "com_github_nats_io_nkeys",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/nats-io/nkeys",
        sum = "h1:6JrEfig+HzTH85yxzhSVbjHRJv9cn0p6n3IngIcM5/k=",
        version = "v0.1.3",
    )
    go_repository(
        name = "com_github_nats_io_nuid",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/nats-io/nuid",
        sum = "h1:5iA8DT8V7q8WK2EScv2padNa/rTESc1KdnPw4TC2paw=",
        version = "v1.0.1",
    )
    go_repository(
        name = "com_github_nbio_st",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/nbio/st",
        sum = "h1:W6apQkHrMkS0Muv8G/TipAy/FJl/rCYT0+EuS8+Z0z4=",
        version = "v0.0.0-20140626010706-e9e8d9816f32",
    )
    go_repository(
        name = "com_github_ncw_swift",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/ncw/swift",
        sum = "h1:4DQRPj35Y41WogBxyhOXlrI37nzGlyEcsforeudyYPQ=",
        version = "v1.0.47",
    )
    go_repository(
        name = "com_github_nwaples_rardecode",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/nwaples/rardecode",
        sum = "h1:r7vGuS5akxOnR4JQSkko62RJ1ReCMXxQRPtxsiFMBOs=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_octago_sflags",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/octago/sflags",
        sum = "h1:XceYzkRXGAHa/lSFmKLcaxSrsh4MTuOMQdIGsUD0wlk=",
        version = "v0.2.0",
    )
    go_repository(
        name = "com_github_oklog_ulid",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/oklog/ulid",
        sum = "h1:EGfNDEx6MqHz8B3uNV6QAib1UR2Lm97sHi3ocA6ESJ4=",
        version = "v1.3.1",
    )
    go_repository(
        name = "com_github_oneofone_xxhash",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/OneOfOne/xxhash",
        sum = "h1:KMrpdQIwFcEqXDklaen+P1axHaj9BSKzvpUUfnHldSE=",
        version = "v1.2.2",
    )
    go_repository(
        name = "com_github_opencontainers_go_digest",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/opencontainers/go-digest",
        sum = "h1:WzifXhOVOEOuFYOJAW6aQqW0TooG2iki3E3Ii+WN7gQ=",
        version = "v1.0.0-rc1",
    )
    go_repository(
        name = "com_github_opencontainers_image_spec",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/opencontainers/image-spec",
        sum = "h1:JMemWkRwHx4Zj+fVxWoMCFm/8sYGGrUVojFA6h/TRcI=",
        version = "v1.0.1",
    )
    go_repository(
        name = "com_github_opencontainers_runc",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/opencontainers/runc",
        sum = "h1:/k06BMULKF5hidyoZymkoDCzdJzltZpz/UU4LguQVtc=",
        version = "v1.0.0-rc9",
    )
    go_repository(
        name = "com_github_opencontainers_runtime_spec",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/opencontainers/runtime-spec",
        sum = "h1:eNUVfm/RFLIi1G7flU5/ZRTHvd4kcVuzfRnL6OFlzCI=",
        version = "v0.1.2-0.20190507144316-5b71a03e2700",
    )
    go_repository(
        name = "com_github_opencontainers_runtime_tools",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/opencontainers/runtime-tools",
        sum = "h1:H7DMc6FAjgwZZi8BRqjrAAHWoqEr5e5L6pS4V0ezet4=",
        version = "v0.0.0-20181011054405-1d69bd0f9c39",
    )
    go_repository(
        name = "com_github_openzipkin_zipkin_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/openzipkin/zipkin-go",
        sum = "h1:33/f6xXB6YlOQ9tgTsXVOkdLCJsHTcZJnMy4DnSd6FU=",
        version = "v0.2.0",
    )
    go_repository(
        name = "com_github_otiai10_copy",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/otiai10/copy",
        sum = "h1:DDNipYy6RkIkjMwy+AWzgKiNTyj2RUI9yEMeETEpVyc=",
        version = "v1.0.2",
    )
    go_repository(
        name = "com_github_otiai10_curr",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/otiai10/curr",
        sum = "h1:+OLn68pqasWca0z5ryit9KGfp3sUsW4Lqg32iRMJyzs=",
        version = "v0.0.0-20150429015615-9b4961190c95",
    )
    go_repository(
        name = "com_github_otiai10_mint",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/otiai10/mint",
        sum = "h1:Ady6MKVezQwHBkGzLFbrsywyp09Ah7rkmfjV3Bcr5uc=",
        version = "v1.3.0",
    )
    go_repository(
        name = "com_github_pelletier_go_buffruneio",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/pelletier/go-buffruneio",
        sum = "h1:U4t4R6YkofJ5xHm3dJzuRpPZ0mr5MMCoAWooScCR7aA=",
        version = "v0.2.0",
    )
    go_repository(
        name = "com_github_phayes_freeport",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/phayes/freeport",
        sum = "h1:JhzVVoYvbOACxoUmOs6V/G4D5nPVUW73rKvXxP4XUJc=",
        version = "v0.0.0-20180830031419-95f893ade6f2",
    )
    go_repository(
        name = "com_github_pierrec_lz4",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/pierrec/lz4",
        sum = "h1:mFe7ttWaflA46Mhqh+jUfjp2qTbPYxLB2/OyBppH9dg=",
        version = "v2.4.1+incompatible",
    )
    go_repository(
        name = "com_github_pkg_profile",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/pkg/profile",
        sum = "h1:F++O52m40owAmADcojzM+9gyjmMOY/T4oYJkgFDH8RE=",
        version = "v1.2.1",
    )
    go_repository(
        name = "com_github_prometheus_tsdb",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/prometheus/tsdb",
        sum = "h1:YZcsG11NqnK4czYLrWd9mpEuAJIHVQLwdrleYfszMAA=",
        version = "v0.7.1",
    )
    go_repository(
        name = "com_github_rcrowley_go_metrics",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/rcrowley/go-metrics",
        sum = "h1:dY6ETXrvDG7Sa4vE8ZQG4yqWg6UnOcbqTAahkV813vQ=",
        version = "v0.0.0-20190826022208-cac0b30c2563",
    )
    go_repository(
        name = "com_github_rogpeppe_go_internal",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/rogpeppe/go-internal",
        sum = "h1:XU784Pr0wdahMY2bYcyK6N1KuaRAdLtqD4qd8D18Bfs=",
        version = "v1.3.2",
    )
    go_repository(
        name = "com_github_rubiojr_go_vhd",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/rubiojr/go-vhd",
        sum = "h1:ht7N4d/B7Ezf58nvMNVF3OlvDlz9pp+WHVcRNS0nink=",
        version = "v0.0.0-20160810183302-0bfd3b39853c",
    )
    go_repository(
        name = "com_github_russross_blackfriday_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/russross/blackfriday/v2",
        sum = "h1:lPqVAte+HuHNfhJ/0LC98ESWRz8afy9tM/0RK8m9o+Q=",
        version = "v2.0.1",
    )
    go_repository(
        name = "com_github_satori_go_uuid",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/satori/go.uuid",
        sum = "h1:0uYX9dsZ2yD7q2RtLRtPSdGDWzjeM3TbMJP9utgA0ww=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_sclevine_spec",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/sclevine/spec",
        sum = "h1:1Jwdf9jSfDl9NVmt8ndHqbTZ7XCCPbh1jI3hkDBHVYA=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_shopify_logrus_bugsnag",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/Shopify/logrus-bugsnag",
        sum = "h1:UrqY+r/OJnIp5u0s1SbQ8dVfLCZJsnvazdBP5hS4iRs=",
        version = "v0.0.0-20171204204709-577dee27f20d",
    )
    go_repository(
        name = "com_github_shopify_sarama",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/Shopify/sarama",
        sum = "h1:XxJBCZEoWJtoWjf/xRbmGUpAmTZGnuuF0ON0EvxxBrs=",
        version = "v1.23.1",
    )
    go_repository(
        name = "com_github_shopify_toxiproxy",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/Shopify/toxiproxy",
        sum = "h1:TKdv8HiTLgE5wdJuEML90aBgNWsokNbMijUGhmcoBJc=",
        version = "v2.1.4+incompatible",
    )
    go_repository(
        name = "com_github_shurcool_githubv4",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/shurcooL/githubv4",
        sum = "h1:Cocq9/ZZxCoiybhygOR7hX4E3/PkV8eNbd1AEcUvaHM=",
        version = "v0.0.0-20191102174205-af46314aec7b",
    )
    go_repository(
        name = "com_github_shurcool_graphql",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/shurcooL/graphql",
        sum = "h1:tygelZueB1EtXkPI6mQ4o9DQ0+FKW41hTbunoXZCTqk=",
        version = "v0.0.0-20181231061246-d48a9a75455f",
    )
    go_repository(
        name = "com_github_shurcool_sanitized_anchor_name",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/shurcooL/sanitized_anchor_name",
        sum = "h1:PdmoCO6wvbs+7yrJyMORt4/BmY5IYyJwS/kOiWx8mHo=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_smartystreets_assertions",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/smartystreets/assertions",
        sum = "h1:zE9ykElWQ6/NYmHa3jpm/yHnI4xSofP+UP6SpjHcSeM=",
        version = "v0.0.0-20180927180507-b2de0cb4f26d",
    )
    go_repository(
        name = "com_github_smartystreets_goconvey",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/smartystreets/goconvey",
        sum = "h1:fv0U8FUIMPNf1L9lnHLvLhgicrIVChEkdzIKYqbNC9s=",
        version = "v1.6.4",
    )
    go_repository(
        name = "com_github_spaolacci_murmur3",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/spaolacci/murmur3",
        sum = "h1:qLC7fQah7D6K1B0ujays3HV9gkFtllcxhzImRR7ArPQ=",
        version = "v0.0.0-20180118202830-f09979ecbc72",
    )
    go_repository(
        name = "com_github_src_d_gcfg",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/src-d/gcfg",
        sum = "h1:xXbNR5AlLSA315x2UO+fTSSAXCDf+Ar38/6oyGbDKQ4=",
        version = "v1.4.0",
    )
    go_repository(
        name = "com_github_streadway_amqp",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/streadway/amqp",
        sum = "h1:0ngsPmuP6XIjiFRNFYlvKwSr5zff2v+uPHaffZ6/M4k=",
        version = "v0.0.0-20190404075320-75d898a42a94",
    )
    go_repository(
        name = "com_github_subosito_gotenv",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/subosito/gotenv",
        sum = "h1:Slr1R9HxAlEKefgq5jn9U+DnETlIUa6HfgEzj0g5d7s=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_syndtr_gocapability",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/syndtr/gocapability",
        sum = "h1:zLV6q4e8Jv9EHjNg/iHfzwDkCve6Ua5jCygptrtXHvI=",
        version = "v0.0.0-20170704070218-db04d3cc01c8",
    )
    go_repository(
        name = "com_github_tektoncd_pipeline",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/tektoncd/pipeline",
        sum = "h1:kGeWm53R5ggajD/L2KU8kcsZ2lVd4ruN3kdqK1A/NwQ=",
        version = "v0.11.0",
    )
    go_repository(
        name = "com_github_tektoncd_plumbing",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/tektoncd/plumbing",
        sum = "h1:BksmpUwtap3THXJ8Z4KGcotsvpRdFQKySjDHgtc22lA=",
        version = "v0.0.0-20200217163359-cd0db6e567d2",
    )
    go_repository(
        name = "com_github_tektoncd_plumbing_pipelinerun_logs",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/tektoncd/plumbing/pipelinerun-logs",
        sum = "h1:9qeyrQsoPZbHOyOPt0OeB1TCYXfYb5swrxlFWzTIYYk=",
        version = "v0.0.0-20191206114338-712d544c2c21",
    )
    go_repository(
        name = "com_github_ugorji_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/ugorji/go",
        sum = "h1:j4s+tAvLfL3bZyefP2SEWmhBzmuIlH/eqNuPdFPgngw=",
        version = "v1.1.4",
    )
    go_repository(
        name = "com_github_ulikunitz_xz",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/ulikunitz/xz",
        sum = "h1:jGHAfXawEGZQ3blwU5wnWKQJvAraT7Ftq9EXjnXYgt8=",
        version = "v0.5.6",
    )
    go_repository(
        name = "com_github_valyala_bytebufferpool",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/valyala/bytebufferpool",
        sum = "h1:GqA5TC/0021Y/b9FG4Oi9Mr3q7XYx6KllzawFIhcdPw=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_vdemeester_k8s_pkg_credentialprovider",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/vdemeester/k8s-pkg-credentialprovider",
        sum = "h1:IBEhRIcu5HP+Pkhzn9E9z3wV0tp3TFjDkiAQtX2FXFM=",
        version = "v1.13.12-1",
    )
    go_repository(
        name = "com_github_vmware_govmomi",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/vmware/govmomi",
        sum = "h1:gpw/0Ku+6RgF3jsi7fnCLmlcikBHfKBCUcu1qgc16OU=",
        version = "v0.20.3",
    )
    go_repository(
        name = "com_github_xanzy_ssh_agent",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/xanzy/ssh-agent",
        sum = "h1:TCbipTQL2JiiCprBWx9frJ2eJlCYT00NmctrHxVAr70=",
        version = "v0.2.1",
    )
    go_repository(
        name = "com_github_xdg_scram",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/xdg/scram",
        sum = "h1:u40Z8hqBAAQyv+vATcGgV0YCnDjqSL7/q/JyPhhJSPk=",
        version = "v0.0.0-20180814205039-7eeb5667e42c",
    )
    go_repository(
        name = "com_github_xdg_stringprep",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/xdg/stringprep",
        sum = "h1:d9X0esnoa3dFsV0FG35rAT0RIhYFlPq7MiP+DW89La0=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_xeipuuv_gojsonpointer",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/xeipuuv/gojsonpointer",
        sum = "h1:J9EGpcZtP0E/raorCMxlFGSTBrsSlaDGf3jU/qvAE2c=",
        version = "v0.0.0-20180127040702-4e3ac2762d5f",
    )
    go_repository(
        name = "com_github_xeipuuv_gojsonreference",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/xeipuuv/gojsonreference",
        sum = "h1:EzJWgHovont7NscjpAxXsDA8S8BMYve8Y5+7cuRE7R0=",
        version = "v0.0.0-20180127040603-bd5ef7bd5415",
    )
    go_repository(
        name = "com_github_xeipuuv_gojsonschema",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/xeipuuv/gojsonschema",
        sum = "h1:LhYJRs+L4fBtjZUfuSZIKGeVu0QRy8e5Xi7D17UxZ74=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_xi2_xz",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/xi2/xz",
        sum = "h1:nIPpBwaJSVYIxUFsDv3M8ofmx9yWTog9BfvIu0q41lo=",
        version = "v0.0.0-20171230120015-48954b6210f8",
    )
    go_repository(
        name = "com_github_xlab_handysort",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/xlab/handysort",
        sum = "h1:j2hhcujLRHAg872RWAV5yaUrEjHEObwDv3aImCaNLek=",
        version = "v0.0.0-20150421192137-fb3537ed64a1",
    )
    go_repository(
        name = "com_github_yvasiyarov_go_metrics",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/yvasiyarov/go-metrics",
        sum = "h1:+lm10QQTNSBd8DVTNGHx7o/IKu9HYDvLMffDhbyLccI=",
        version = "v0.0.0-20140926110328-57bccd1ccd43",
    )
    go_repository(
        name = "com_github_yvasiyarov_gorelic",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/yvasiyarov/gorelic",
        sum = "h1:hlE8//ciYMztlGpl/VA+Zm1AcTPHYkHJPbHqE6WJUXE=",
        version = "v0.0.0-20141212073537-a9bba5b9ab50",
    )
    go_repository(
        name = "com_github_yvasiyarov_newrelic_platform_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/yvasiyarov/newrelic_platform_go",
        sum = "h1:ERexzlUfuTvpE74urLSbIQW0Z/6hF9t8U4NsJLaioAY=",
        version = "v0.0.0-20140908184405-b21fdbd4370f",
    )
    go_repository(
        name = "com_google_cloud_go_bigquery",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "cloud.google.com/go/bigquery",
        sum = "h1:hL+ycaJpVE9M7nLoiXb/Pn10ENE2u+oddxbD8uu0ZVU=",
        version = "v1.0.1",
    )
    go_repository(
        name = "com_google_cloud_go_datastore",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "cloud.google.com/go/datastore",
        sum = "h1:Kt+gOPPp2LEPWp8CSfxhsM8ik9CcyE/gYu+0r+RnZvM=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_google_cloud_go_logging",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "cloud.google.com/go/logging",
        sum = "h1:kaunpnoEh9L4hu6JUsBa8Y20LBfKnCuDhKUgdZp7oK8=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_google_cloud_go_pubsub",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "cloud.google.com/go/pubsub",
        sum = "h1:W9tAK3E57P75u0XLLR82LZyw8VpAnhmyTOxW9qzmyj8=",
        version = "v1.0.1",
    )
    go_repository(
        name = "com_google_cloud_go_storage",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "cloud.google.com/go/storage",
        sum = "h1:VV2nUM3wwLLGh9lSABFgZMjInyUbJeaRSE64WuAIQ+4=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_shuralyov_dmitri_gpu_mtl",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "dmitri.shuralyov.com/gpu/mtl",
        sum = "h1:VpgP7xuJadIUuKccphEpTJnWhS2jkQyMt6Y7pJCD7fY=",
        version = "v0.0.0-20190408044501-666a987793e9",
    )
    go_repository(
        name = "dev_gocloud",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "gocloud.dev",
        sum = "h1:EDRyaRAnMGSq/QBto486gWFxMLczAfIYUmusV7XLNBM=",
        version = "v0.19.0",
    )
    go_repository(
        name = "dev_knative_caching",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "knative.dev/caching",
        sum = "h1:mxrur6DsVK8uIjhIq7c1OMls4YjBcRlyvnh3Vx13a0M=",
        version = "v0.0.0-20200116200605-67bca2c83dfa",
    )
    go_repository(
        name = "dev_knative_eventing_contrib",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "knative.dev/eventing-contrib",
        sum = "h1:xncT+JrokPG+hPUFJwue8ubPpzmziV9GUIZqYt01JDo=",
        version = "v0.11.2",
    )
    go_repository(
        name = "dev_knative_pkg",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "knative.dev/pkg",
        sum = "h1:52b67wiu9B62n+ZsDAMjHt84sZfiR0CUBTvtF1UEGmo=",
        version = "v0.0.0-20200207155214-fef852970f43",
    )
    go_repository(
        name = "in_gopkg_airbrake_gobrake_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "gopkg.in/airbrake/gobrake.v2",
        sum = "h1:7z2uVWwn7oVeeugY1DtlPAy5H+KYgB1KeKTnqjNatLo=",
        version = "v2.0.9",
    )
    go_repository(
        name = "in_gopkg_errgo_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "gopkg.in/errgo.v2",
        sum = "h1:0vLT13EuvQ0hNvakwLuFZ/jYrLp5F3kcWHXdRggjCE8=",
        version = "v2.1.0",
    )
    go_repository(
        name = "in_gopkg_gcfg_v1",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "gopkg.in/gcfg.v1",
        sum = "h1:0HIbH907iBTAntm+88IJV2qmJALDAh8sPekI9Vc1fm0=",
        version = "v1.2.0",
    )
    go_repository(
        name = "in_gopkg_gemnasium_logrus_airbrake_hook_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "gopkg.in/gemnasium/logrus-airbrake-hook.v2",
        sum = "h1:OAj3g0cR6Dx/R07QgQe8wkA9RNjB2u4i700xBkIT4e0=",
        version = "v2.1.2",
    )
    go_repository(
        name = "in_gopkg_ini_v1",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "gopkg.in/ini.v1",
        sum = "h1:j+Lt/M1oPPejkniCg1TkWE2J3Eh1oZTsHSXzMTzUXn4=",
        version = "v1.52.0",
    )
    go_repository(
        name = "in_gopkg_jcmturner_aescts_v1",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "gopkg.in/jcmturner/aescts.v1",
        sum = "h1:cVVZBK2b1zY26haWB4vbBiZrfFQnfbTVrE3xZq6hrEw=",
        version = "v1.0.1",
    )
    go_repository(
        name = "in_gopkg_jcmturner_dnsutils_v1",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "gopkg.in/jcmturner/dnsutils.v1",
        sum = "h1:cIuC1OLRGZrld+16ZJvvZxVJeKPsvd5eUIvxfoN5hSM=",
        version = "v1.0.1",
    )
    go_repository(
        name = "in_gopkg_jcmturner_gokrb5_v7",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "gopkg.in/jcmturner/gokrb5.v7",
        sum = "h1:a9tsXlIDD9SKxotJMK3niV7rPZAJeX2aD/0yg3qlIrg=",
        version = "v7.5.0",
    )
    go_repository(
        name = "in_gopkg_jcmturner_rpc_v1",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "gopkg.in/jcmturner/rpc.v1",
        sum = "h1:QHIUxTX1ISuAv9dD2wJ9HWQVuWDX/Zc0PfeC2tjc4rU=",
        version = "v1.1.0",
    )
    go_repository(
        name = "in_gopkg_robfig_cron_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "gopkg.in/robfig/cron.v2",
        sum = "h1:E846t8CnR+lv5nE+VuiKTDG/v1U2stad0QzddfJC7kY=",
        version = "v2.0.0-20150107220207-be2e0b0deed5",
    )
    go_repository(
        name = "in_gopkg_src_d_go_billy_v4",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "gopkg.in/src-d/go-billy.v4",
        sum = "h1:0SQA1pRztfTFx2miS8sA97XvooFeNOmvUenF4o0EcVg=",
        version = "v4.3.2",
    )
    go_repository(
        name = "in_gopkg_src_d_go_git_fixtures_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "gopkg.in/src-d/go-git-fixtures.v3",
        sum = "h1:ivZFOIltbce2Mo8IjzUHAFoq/IylO9WHhNOAJK+LsJg=",
        version = "v3.5.0",
    )
    go_repository(
        name = "in_gopkg_src_d_go_git_v4",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "gopkg.in/src-d/go-git.v4",
        sum = "h1:SRtFyV8Kxc0UP7aCHcijOMQGPxHSmMOPrzulQWolkYE=",
        version = "v4.13.1",
    )
    go_repository(
        name = "in_gopkg_warnings_v0",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "gopkg.in/warnings.v0",
        sum = "h1:wFXVbFY8DY5/xOe1ECiWdKCzZlxgshcYVNkBHstARME=",
        version = "v0.1.2",
    )
    go_repository(
        name = "io_k8s_cli_runtime",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "k8s.io/cli-runtime",
        sum = "h1:0ZlDdJgJBKsu77trRUynNiWsRuAvAVPBNaQfnt/1qtc=",
        version = "v0.17.3",
    )
    go_repository(
        name = "io_k8s_cloud_provider",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "k8s.io/cloud-provider",
        sum = "h1:BQZPD1Ja/vnTOj1GKI9/wSpd3qgIDZp9q2NAS3568Ac=",
        version = "v0.17.0",
    )
    go_repository(
        name = "io_k8s_csi_translation_lib",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "k8s.io/csi-translation-lib",
        sum = "h1:8hwWJDMOBCAogaWXtNWy0dYGQ2dZYzOnOzjQMiDaY+E=",
        version = "v0.17.0",
    )
    go_repository(
        name = "io_k8s_kubectl",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "k8s.io/kubectl",
        sum = "h1:QZR8Q6lWiVRjwKslekdbN5WPMp53dS/17j5e+oi5XVU=",
        version = "v0.17.2",
    )
    go_repository(
        name = "io_k8s_kubernetes",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "k8s.io/kubernetes",
        sum = "h1:qTfB+u5M92k2fCCCVP2iuhgwwSOv1EkAkvQY1tQODD8=",
        version = "v1.13.0",
    )
    go_repository(
        name = "io_k8s_legacy_cloud_providers",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "k8s.io/legacy-cloud-providers",
        sum = "h1:ITm7sUthpxQyP96MU7K4Ra9M9M1k9eywUWv9IiTaxzc=",
        version = "v0.17.0",
    )
    go_repository(
        name = "io_k8s_metrics",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "k8s.io/metrics",
        sum = "h1:cuN1ScyUS9/tj4YFI8d0/7yO0BveFHhyQpPNWS8uLr8=",
        version = "v0.17.2",
    )
    go_repository(
        name = "io_k8s_sigs_apiserver_network_proxy_konnectivity_client",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "sigs.k8s.io/apiserver-network-proxy/konnectivity-client",
        sum = "h1:uuHDyjllyzRyCIvvn0OBjiRB0SgBZGqHNYAmjR7fO50=",
        version = "v0.0.7",
    )
    go_repository(
        name = "io_k8s_sigs_boskos",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "sigs.k8s.io/boskos",
        sum = "h1:O/XIKGebDWPSGEkC7eCeoacAEeJjoqLMFxlq5BuS5tI=",
        version = "v0.0.0-20200710214748-f5935686c7fc",
    )
    go_repository(
        name = "io_k8s_sigs_kubetest2",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "sigs.k8s.io/kubetest2",
        sum = "h1:GaAGMxDikZzx2xF5io1nBNTrdc4HSxutKGvu2g7BbCQ=",
        version = "v0.0.0-20200807173356-3d574132ed2e",
    )
    go_repository(
        name = "io_k8s_sigs_kustomize",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "sigs.k8s.io/kustomize",
        sum = "h1:JUufWFNlI44MdtnjUqVnvh29rR37PQFzPbLXqhyOyX0=",
        version = "v2.0.3+incompatible",
    )
    go_repository(
        name = "io_k8s_test_infra",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "k8s.io/test-infra",
        sum = "h1:1moyKLm99rwMKnGQrFjOfGo+lAOQnC+jysslmVvw1FI=",
        version = "v0.0.0-20200617221206-ea73eaeab7ff",
    )
    go_repository(
        name = "io_opencensus_go_contrib_exporter_aws",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "contrib.go.opencensus.io/exporter/aws",
        sum = "h1:YsbWYxDZkC7x2OxlsDEYvvEXZ3cBI3qBgUK5BqkZvRw=",
        version = "v0.0.0-20181029163544-2befc13012d0",
    )
    go_repository(
        name = "io_opencensus_go_contrib_exporter_ocagent",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "contrib.go.opencensus.io/exporter/ocagent",
        sum = "h1:TKXjQSRS0/cCDrP7KvkgU6SmILtF/yV2TOs/02K/WZQ=",
        version = "v0.5.0",
    )
    go_repository(
        name = "io_opencensus_go_contrib_exporter_prometheus",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "contrib.go.opencensus.io/exporter/prometheus",
        sum = "h1:SByaIoWwNgMdPSgl5sMqM2KDE5H/ukPWBRo314xiDvg=",
        version = "v0.1.0",
    )
    go_repository(
        name = "io_opencensus_go_contrib_exporter_stackdriver",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "contrib.go.opencensus.io/exporter/stackdriver",
        sum = "h1:iXI5hr7pUwMx0IwMphpKz5Q3If/G5JiWFVZ5MPPxP9E=",
        version = "v0.12.8",
    )
    go_repository(
        name = "io_opencensus_go_contrib_integrations_ocsql",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "contrib.go.opencensus.io/integrations/ocsql",
        sum = "h1:kfg5Yyy1nYUrqzyfW5XX+dzMASky8IJXhtHe0KTYNS4=",
        version = "v0.1.4",
    )
    go_repository(
        name = "io_opencensus_go_contrib_resource",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "contrib.go.opencensus.io/resource",
        sum = "h1:4r2CANuYhKGmYWP02+5E94rLRcS/YeD+KlxSrOsMxk0=",
        version = "v0.1.1",
    )
    go_repository(
        name = "io_rsc_binaryregexp",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "rsc.io/binaryregexp",
        sum = "h1:HfqmD5MEmC0zvwBuF187nq9mdnXjXsSivRiXN7SmRkE=",
        version = "v0.2.0",
    )
    go_repository(
        name = "io_rsc_letsencrypt",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "rsc.io/letsencrypt",
        sum = "h1:H7xDfhkaFFSYEJlKeq38RwX2jYcnTeHuDQyT+mMNMwM=",
        version = "v0.0.3",
    )
    go_repository(
        name = "ml_vbom_util",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "vbom.ml/util",
        sum = "h1:O69FD9pJA4WUZlEwYatBEEkRWKQ5cKodWpdKTrCS/iQ=",
        version = "v0.0.0-20180919145318-efcd4e0f9787",
    )
    go_repository(
        name = "org_apache_git_thrift_git",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "git.apache.org/thrift.git",
        sum = "h1:CMxsZlAmxKs+VAZMlDDL0wXciMblJcutQbEe3A9CYUM=",
        version = "v0.12.0",
    )
    go_repository(
        name = "org_bazil_fuse",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "bazil.org/fuse",
        sum = "h1:FNCRpXiquG1aoyqcIWVFmpTSKVcx2bQD38uZZeGtdlw=",
        version = "v0.0.0-20180421153158-65cc252bf669",
    )
    go_repository(
        name = "org_golang_google_cloud",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "google.golang.org/cloud",
        sum = "h1:Cpp2P6TPjujNoC5M2KHY6g7wfyLYfIWRZaSdIKfDasA=",
        version = "v0.0.0-20151119220103-975617b05ea8",
    )
    go_repository(
        name = "org_golang_x_mod",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "golang.org/x/mod",
        sum = "h1:RM4zey1++hCTbCVQfnWeKs9/IEsaBLA8vTkd0WVtmH4=",
        version = "v0.3.0",
    )
    go_repository(
        name = "org_uber_go_tools",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "go.uber.org/tools",
        sum = "h1:0mgffUl7nfd+FpvXMVz4IDEaUSmT1ysygQC7qYo7sG4=",
        version = "v0.0.0-20190618225709-2cfd321de3ee",
    )
    go_repository(
        name = "sh_helm_helm_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "helm.sh/helm/v3",
        sum = "h1:aykwPMVyQyncZ8iLNVMXgJ1l3c6W0+LSOPmqp8JdCjs=",
        version = "v3.1.1",
    )
    go_repository(
        name = "com_github_cenkalti_backoff",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/cenkalti/backoff",
        sum = "h1:tNowT99t7UNflLxfYYSlKYsBpXdEet03Pg2g16Swow4=",
        version = "v2.2.1+incompatible",
    )
    go_repository(
        name = "com_github_ajg_form",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/ajg/form",
        sum = "h1:t9c7v8JUKu/XxOGBU0yjNpaMloxGEJhUkqFRq0ibGeU=",
        version = "v1.5.1",
    )
    go_repository(
        name = "com_github_andreasbriese_bbloom",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/AndreasBriese/bbloom",
        sum = "h1:HD8gA2tkByhMAwYaFAX9w2l7vxvBQ5NMoxDrkhqhtn4=",
        version = "v0.0.0-20190306092124-e2d15f34fcf9",
    )
    go_repository(
        name = "com_github_aymerick_raymond",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/aymerick/raymond",
        sum = "h1:Ppm0npCCsmuR9oQaBtRuZcmILVE74aXE+AmrJj8L2ns=",
        version = "v2.0.3-0.20180322193309-b565731e1464+incompatible",
    )
    go_repository(
        name = "com_github_cloudykit_fastprinter",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/CloudyKit/fastprinter",
        sum = "h1:3SgJcK9l5uPdBC/X17wanyJAMxM33+4ZhEIV96MIH8U=",
        version = "v0.0.0-20170127035650-74b38d55f37a",
    )
    go_repository(
        name = "com_github_cloudykit_jet",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/CloudyKit/jet",
        sum = "h1:rZgFj+Gtf3NMi/U5FvCvhzaxzW/TaPYgUYx3bAPz9DE=",
        version = "v2.1.3-0.20180809161101-62edd43e4f88+incompatible",
    )
    go_repository(
        name = "com_github_cncf_udpa_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/cncf/udpa/go",
        sum = "h1:WBZRG4aNOuI15bLRrCgN8fCq8E5Xuty6jGbmSNEvSsU=",
        version = "v0.0.0-20191209042840-269d4d468f6f",
    )
    go_repository(
        name = "com_github_cockroachdb_errors",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/cockroachdb/errors",
        sum = "h1:4IrYIc17U7TSuLYlol83tc7ZKmJIs8PbJ/YE+bzoyik=",
        version = "v1.8.0",
    )
    go_repository(
        name = "com_github_cockroachdb_logtags",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/cockroachdb/logtags",
        sum = "h1:o/kfcElHqOiXqcou5a3rIlMc7oJbMQkeLk0VQJ7zgqY=",
        version = "v0.0.0-20190617123548-eb05cc24525f",
    )
    go_repository(
        name = "com_github_cockroachdb_redact",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/cockroachdb/redact",
        sum = "h1:7cjFsEgqTuNh+72+gtpnVTCqTiG0vT86ffETvTxyvUo=",
        version = "v1.0.7",
    )
    go_repository(
        name = "com_github_cockroachdb_sentry_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/cockroachdb/sentry-go",
        sum = "h1:IKgmqgMQlVJIZj19CdocBeSfSaiCbEBZGKODaixqtHM=",
        version = "v0.6.1-cockroachdb.2",
    )
    go_repository(
        name = "com_github_codegangsta_inject",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/codegangsta/inject",
        sum = "h1:sDMmm+q/3+BukdIpxwO365v/Rbspp2Nt5XntgQRXq8Q=",
        version = "v0.0.0-20150114235600-33e0aa1cb7c0",
    )
    go_repository(
        name = "com_github_dgraph_io_badger",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/dgraph-io/badger",
        sum = "h1:DshxFxZWXUcO0xX476VJC07Xsr6ZCBVRHKZ93Oh7Evo=",
        version = "v1.6.0",
    )
    go_repository(
        name = "com_github_dgryski_go_farm",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/dgryski/go-farm",
        sum = "h1:tdlZCpZ/P9DhczCTSixgIKmwPv6+wP5DGjqLYw5SUiA=",
        version = "v0.0.0-20190423205320-6a90982ecee2",
    )
    go_repository(
        name = "com_github_eknkc_amber",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/eknkc/amber",
        sum = "h1:clC1lXBpe2kTj2VHdaIu9ajZQe4kcEY9j0NsnDDBZ3o=",
        version = "v0.0.0-20171010120322-cdade1c07385",
    )
    go_repository(
        name = "com_github_etcd_io_bbolt",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/etcd-io/bbolt",
        sum = "h1:gSJmxrs37LgTqR/oyJBWok6k6SvXEUerFTbltIhXkBM=",
        version = "v1.3.3",
    )
    go_repository(
        name = "com_github_fasthttp_contrib_websocket",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/fasthttp-contrib/websocket",
        sum = "h1:DddqAaWDpywytcG8w/qoQ5sAN8X12d3Z3koB0C3Rxsc=",
        version = "v0.0.0-20160511215533-1f3b11f56072",
    )
    go_repository(
        name = "com_github_fatih_structs",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/fatih/structs",
        sum = "h1:Q7juDM0QtcnhCpeyLGQKyg4TOIghuNXrkL32pHAUMxo=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_flosch_pongo2",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/flosch/pongo2",
        sum = "h1:GY1+t5Dr9OKADM64SYnQjw/w99HMYvQ0A8/JoUkxVmc=",
        version = "v0.0.0-20190707114632-bbf5a6c351f4",
    )
    go_repository(
        name = "com_github_gavv_httpexpect",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/gavv/httpexpect",
        sum = "h1:1X9kcRshkSKEjNJJxX9Y9mQ5BRfbxU5kORdjhlA1yX8=",
        version = "v2.0.0+incompatible",
    )
    go_repository(
        name = "com_github_gin_contrib_sse",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/gin-contrib/sse",
        sum = "h1:t8FVkw33L+wilf2QiWkw0UV77qRpcH/JHPKGpKa2E8g=",
        version = "v0.0.0-20190301062529-5545eab6dad3",
    )
    go_repository(
        name = "com_github_gin_gonic_gin",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/gin-gonic/gin",
        sum = "h1:3tMoCCfM7ppqsR0ptz/wi1impNpT7/9wQtMZ8lr1mCQ=",
        version = "v1.4.0",
    )
    go_repository(
        name = "com_github_go_check_check",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/go-check/check",
        sum = "h1:0gkP6mzaMqkmpcJYCFOLkIBwI7xFExG03bbkOkCvUPI=",
        version = "v0.0.0-20180628173108-788fd7840127",
    )
    go_repository(
        name = "com_github_go_errors_errors",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/go-errors/errors",
        sum = "h1:LUHzmkK3GUKUrL/1gfBUxAHzcev3apQlezX/+O7ma6w=",
        version = "v1.0.1",
    )
    go_repository(
        name = "com_github_go_martini_martini",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/go-martini/martini",
        sum = "h1:xveKWz2iaueeTaUgdetzel+U7exyigDYBryyVfV/rZk=",
        version = "v0.0.0-20170121215854-22fa46961aab",
    )
    go_repository(
        name = "com_github_gobwas_httphead",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/gobwas/httphead",
        sum = "h1:s+21KNqlpePfkah2I+gwHF8xmJWRjooY+5248k6m4A0=",
        version = "v0.0.0-20180130184737-2c6c146eadee",
    )
    go_repository(
        name = "com_github_gobwas_pool",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/gobwas/pool",
        sum = "h1:QEmUOlnSjWtnpRGHF3SauEiOsy82Cup83Vf2LcMlnc8=",
        version = "v0.2.0",
    )
    go_repository(
        name = "com_github_gobwas_ws",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/gobwas/ws",
        sum = "h1:CoAavW/wd/kulfZmSIBt6p24n4j7tHgNVCjsfHVNUbo=",
        version = "v1.0.2",
    )
    go_repository(
        name = "com_github_gogo_status",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/gogo/status",
        sum = "h1:+eIkrewn5q6b30y+g/BJINVVdi2xH7je5MPJ3ZPK3JA=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_hashicorp_go_version",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/hashicorp/go-version",
        sum = "h1:3vNe/fWF5CBgRIguda1meWhsZHy3m8gCJ5wx+dIzX/E=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_hydrogen18_memlistener",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/hydrogen18/memlistener",
        sum = "h1:EPRgaDqXpLFUJLXZdGLnBTy1l6CLiNAPnvn2l+kHit0=",
        version = "v0.0.0-20141126152155-54553eb933fb",
    )
    go_repository(
        name = "com_github_imkira_go_interpol",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/imkira/go-interpol",
        sum = "h1:KIiKr0VSG2CUW1hl1jpiyuzuJeKUUpC8iM1AIE7N1Vk=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_iris_contrib_blackfriday",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/iris-contrib/blackfriday",
        sum = "h1:o5sHQHHm0ToHUlAJSTjW9UWicjJSDDauOOQ2AHuIVp4=",
        version = "v2.0.0+incompatible",
    )
    go_repository(
        name = "com_github_iris_contrib_go_uuid",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/iris-contrib/go.uuid",
        sum = "h1:XZubAYg61/JwnJNbZilGjf3b3pB80+OQg2qf6c8BfWE=",
        version = "v2.0.0+incompatible",
    )
    go_repository(
        name = "com_github_iris_contrib_i18n",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/iris-contrib/i18n",
        sum = "h1:Kyp9KiXwsyZRTeoNjgVCrWks7D8ht9+kg6yCjh8K97o=",
        version = "v0.0.0-20171121225848-987a633949d0",
    )
    go_repository(
        name = "com_github_iris_contrib_schema",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/iris-contrib/schema",
        sum = "h1:10g/WnoRR+U+XXHWKBHeNy/+tZmM2kcAVGLOsz+yaDA=",
        version = "v0.0.1",
    )
    go_repository(
        name = "com_github_joker_hpp",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/Joker/hpp",
        sum = "h1:65+iuJYdRXv/XyN62C1uEmmOx3432rNG/rKlX6V7Kkc=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_joker_jade",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/Joker/jade",
        sum = "h1:mreN1m/5VJ/Zc3b4pzj9qU6D9SRQ6Vm+3KfI328t3S8=",
        version = "v1.0.1-0.20190614124447-d475f43051e7",
    )
    go_repository(
        name = "com_github_juju_errors",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/juju/errors",
        sum = "h1:rhqTjzJlm7EbkELJDKMTU7udov+Se0xZkWmugr6zGok=",
        version = "v0.0.0-20181118221551-089d3ea4e4d5",
    )
    go_repository(
        name = "com_github_juju_loggo",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/juju/loggo",
        sum = "h1:MK144iBQF9hTSwBW/9eJm034bVoG30IshVm688T2hi8=",
        version = "v0.0.0-20180524022052-584905176618",
    )
    go_repository(
        name = "com_github_juju_testing",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/juju/testing",
        sum = "h1:WQM1NildKThwdP7qWrNAFGzp4ijNLw8RlgENkaI4MJs=",
        version = "v0.0.0-20180920084828-472a3e8b2073",
    )
    go_repository(
        name = "com_github_k0kubun_colorstring",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/k0kubun/colorstring",
        sum = "h1:uC1QfSlInpQF+M0ao65imhwqKnz3Q2z/d8PWZRMQvDM=",
        version = "v0.0.0-20150214042306-9440f1994b88",
    )
    go_repository(
        name = "com_github_kataras_golog",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/kataras/golog",
        sum = "h1:J7Dl82843nbKQDrQM/abbNJZvQjS6PfmkkffhOTXEpM=",
        version = "v0.0.9",
    )
    go_repository(
        name = "com_github_kataras_iris_v12",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/kataras/iris/v12",
        sum = "h1:Wo5S7GMWv5OAzJmvFTvss/C4TS1W0uo6LkDlSymT4rM=",
        version = "v12.0.1",
    )
    go_repository(
        name = "com_github_kataras_neffos",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/kataras/neffos",
        sum = "h1:O06dvQlxjdWvzWbm2Bq+Si6psUhvSmEctAMk9Xujqms=",
        version = "v0.0.10",
    )
    go_repository(
        name = "com_github_kataras_pio",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/kataras/pio",
        sum = "h1:V5Rs9ztEWdp58oayPq/ulmlqJJZeJP6pP79uP3qjcao=",
        version = "v0.0.0-20190103105442-ea782b38602d",
    )
    go_repository(
        name = "com_github_labstack_echo_v4",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/labstack/echo/v4",
        sum = "h1:z0BZoArY4FqdpUEl+wlHp4hnr/oSR6MTmQmv8OHSoww=",
        version = "v4.1.11",
    )
    go_repository(
        name = "com_github_labstack_gommon",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/labstack/gommon",
        sum = "h1:JEeO0bvc78PKdyHxloTKiF8BD5iGrH8T6MSeGvSgob0=",
        version = "v0.3.0",
    )
    go_repository(
        name = "com_github_mattn_goveralls",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/mattn/goveralls",
        sum = "h1:7eJB6EqsPhRVxvwEXGnqdO2sJI0PTsrWoTMXEk9/OQc=",
        version = "v0.0.2",
    )
    go_repository(
        name = "com_github_mediocregopher_mediocre_go_lib",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/mediocregopher/mediocre-go-lib",
        sum = "h1:3dQJqqDouawQgl3gBE1PNHKFkJYGEuFb1DbSlaxdosE=",
        version = "v0.0.0-20181029021733-cb65787f37ed",
    )
    go_repository(
        name = "com_github_mediocregopher_radix_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/mediocregopher/radix/v3",
        sum = "h1:oacPXPKHJg0hcngVVrdtTnfGJiS+PtwoQwTBZGFlV4k=",
        version = "v3.3.0",
    )
    go_repository(
        name = "com_github_microcosm_cc_bluemonday",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/microcosm-cc/bluemonday",
        sum = "h1:5lPfLTTAvAbtS0VqT+94yOtFnGfUWYyx0+iToC3Os3s=",
        version = "v1.0.2",
    )
    go_repository(
        name = "com_github_moul_http2curl",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/moul/http2curl",
        sum = "h1:dRMWoAtb+ePxMlLkrCbAqh4TlPHXvoGUSQ323/9Zahs=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_nxadm_tail",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/nxadm/tail",
        sum = "h1:DQuhQpB1tVlglWS2hLQ5OV6B5r8aGxSrPc5Qo6uTN78=",
        version = "v1.4.4",
    )
    go_repository(
        name = "com_github_pingcap_errors",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/pingcap/errors",
        sum = "h1:lFuQV/oaUMGcD2tqt+01ROSmJs75VG1ToEOkZIZ4nE4=",
        version = "v0.11.4",
    )
    go_repository(
        name = "com_github_ryanuber_columnize",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/ryanuber/columnize",
        sum = "h1:j1Wcmh8OrK4Q7GXY+V7SVSY8nUWQxHW5TkBe7YUl+2s=",
        version = "v2.1.0+incompatible",
    )
    go_repository(
        name = "com_github_sclevine_agouti",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/sclevine/agouti",
        sum = "h1:8IBJS6PWz3uTlMP3YBIR5f+KAldcGuOeFkFbUWfBgK4=",
        version = "v3.0.0+incompatible",
    )
    go_repository(
        name = "com_github_shopify_goreferrer",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/Shopify/goreferrer",
        sum = "h1:WDC6ySpJzbxGWFh4aMxFFC28wwGp5pEuoTtvA4q/qQ4=",
        version = "v0.0.0-20181106222321-ec9c9a553398",
    )
    go_repository(
        name = "com_github_urfave_negroni",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/urfave/negroni",
        sum = "h1:kIimOitoypq34K7TG7DUaJ9kq/N4Ofuwi1sjz0KipXc=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_valyala_fasthttp",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/valyala/fasthttp",
        sum = "h1:uWF8lgKmeaIewWVPwi4GRq2P6+R46IgYZdxWtM+GtEY=",
        version = "v1.6.0",
    )
    go_repository(
        name = "com_github_valyala_fasttemplate",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/valyala/fasttemplate",
        sum = "h1:tY9CJiPnMXf1ERmG2EyK7gNUd+c6RKGD0IfU8WdUSz8=",
        version = "v1.0.1",
    )
    go_repository(
        name = "com_github_valyala_tcplisten",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/valyala/tcplisten",
        sum = "h1:0R4NLDRDZX6JcmhJgXi5E4b8Wg84ihbmUKp/GvSPEzc=",
        version = "v0.0.0-20161114210144-ceec8f93295a",
    )
    go_repository(
        name = "com_github_yalp_jsonpath",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/yalp/jsonpath",
        sum = "h1:6fRhSjgLCkTD3JnJxvaJ4Sj+TYblw757bqYgZaOq5ZY=",
        version = "v0.0.0-20180802001716-5cc68e5049a0",
    )
    go_repository(
        name = "com_github_yudai_gojsondiff",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/yudai/gojsondiff",
        sum = "h1:27cbfqXLVEJ1o8I6v3y9lg8Ydm53EKqHXAOMxEGlCOA=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_yudai_golcs",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/yudai/golcs",
        sum = "h1:BHyfKlQyqbsFN5p3IfnEUduWvb9is428/nNb5L3U01M=",
        version = "v0.0.0-20170316035057-ecda9a501e82",
    )
    go_repository(
        name = "com_github_yudai_pp",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/yudai/pp",
        sum = "h1:Q4//iY4pNF6yPLZIigmvcl7k/bPgrcTPIFIcmawg5bI=",
        version = "v2.0.1+incompatible",
    )
    go_repository(
        name = "in_gopkg_go_playground_assert_v1",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "gopkg.in/go-playground/assert.v1",
        sum = "h1:xoYuJVE7KT85PYWrN730RguIQO0ePzVRfFMXadIrXTM=",
        version = "v1.2.1",
    )
    go_repository(
        name = "in_gopkg_go_playground_validator_v8",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "gopkg.in/go-playground/validator.v8",
        sum = "h1:lFB4DoMU6B626w8ny76MV7VX6W2VHct2GVOI3xgiMrQ=",
        version = "v8.18.2",
    )
    go_repository(
        name = "in_gopkg_mgo_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "gopkg.in/mgo.v2",
        sum = "h1:xcEWjVhvbDy+nHP67nPDDpbYrY+ILlfndk4bRioVHaU=",
        version = "v2.0.0-20180705113604-9856a29383ce",
    )
    go_repository(
        name = "org_golang_google_protobuf",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "google.golang.org/protobuf",
        sum = "h1:4MY060fB1DLGMB/7MBTLnwQUY6+F09GEiz6SsrNqyzM=",
        version = "v1.23.0",
    )
    go_repository(
        name = "com_github_cockroachdb_apd",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/cockroachdb/apd",
        sum = "h1:3LFP3629v+1aKXU5Q37mxmRxX/pIu1nijXydLShEq5I=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_jackc_chunkreader",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/jackc/chunkreader",
        sum = "h1:4s39bBR8ByfqH+DKm8rQA3E1LHZWB9XWcrz8fqaZbe0=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_jackc_chunkreader_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/jackc/chunkreader/v2",
        sum = "h1:i+RDz65UE+mmpjTfyz0MoVTnzeYxroil2G82ki7MGG8=",
        version = "v2.0.1",
    )
    go_repository(
        name = "com_github_jackc_pgconn",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/jackc/pgconn",
        sum = "h1:pwjzcYyfmz/HQOQlENvG1OcDqauTGaqlVahq934F0/U=",
        version = "v1.7.0",
    )
    go_repository(
        name = "com_github_jackc_pgio",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/jackc/pgio",
        sum = "h1:g12B9UwVnzGhueNavwioyEEpAmqMe1E/BN9ES+8ovkE=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_jackc_pgmock",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/jackc/pgmock",
        sum = "h1:JVX6jT/XfzNqIjye4717ITLaNwV9mWbJx0dLCpcRzdA=",
        version = "v0.0.0-20190831213851-13a1b77aafa2",
    )
    go_repository(
        name = "com_github_jackc_pgpassfile",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/jackc/pgpassfile",
        sum = "h1:/6Hmqy13Ss2zCq62VdNG8tM1wchn8zjSGOBJ6icpsIM=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_jackc_pgproto3",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/jackc/pgproto3",
        sum = "h1:FYYE4yRw+AgI8wXIinMlNjBbp/UitDJwfj5LqqewP1A=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_jackc_pgproto3_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/jackc/pgproto3/v2",
        sum = "h1:NUbEWPmCQZbMmYlTjVoNPhc0CfnYyz2bfUAh6A5ZVJM=",
        version = "v2.0.5",
    )
    go_repository(
        name = "com_github_jackc_pgservicefile",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/jackc/pgservicefile",
        sum = "h1:C8S2+VttkHFdOOCXJe+YGfa4vHYwlt4Zx+IVXQ97jYg=",
        version = "v0.0.0-20200714003250-2b9c44734f2b",
    )
    go_repository(
        name = "com_github_jackc_pgtype",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/jackc/pgtype",
        sum = "h1:jzBqRk2HFG2CV4AIwgCI2PwTgm6UUoCAK2ofHHRirtc=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_github_jackc_pgx_v4",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/jackc/pgx/v4",
        sum = "h1:6STjDqppM2ROy5p1wNDcsC7zJTjSHeuCsguZmXyzx7c=",
        version = "v4.9.0",
    )
    go_repository(
        name = "com_github_jackc_puddle",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/jackc/puddle",
        sum = "h1:mpQEXihFnWGDy6X98EOTh81JYuxn7txby8ilJ3iIPGM=",
        version = "v1.1.2",
    )
    go_repository(
        name = "com_github_rs_xid",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/rs/xid",
        sum = "h1:mhH9Nq+C1fY2l1XIpgxIiUOfNpRBYH1kKcr+qfKgjRc=",
        version = "v1.2.1",
    )
    go_repository(
        name = "com_github_rs_zerolog",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/rs/zerolog",
        sum = "h1:uPRuwkWF4J6fGsJ2R0Gn2jB1EQiav9k3S6CSdygQJXY=",
        version = "v1.15.0",
    )
    go_repository(
        name = "com_github_shopspring_decimal",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/shopspring/decimal",
        sum = "h1:abSATXmQEYyShuxI4/vyW3tV1MrKAJzCZ/0zLUXYbsQ=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_zenazn_goji",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/zenazn/goji",
        sum = "h1:RSQQAbXGArQ0dIDEq+PI6WqN6if+5KHu6x2Cx/GXLTQ=",
        version = "v0.9.0",
    )
    go_repository(
        name = "in_gopkg_inconshreveable_log15_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "gopkg.in/inconshreveable/log15.v2",
        sum = "h1:RlWgLqCMMIYYEVcAR5MDsuHlVkaIPDAF+5Dehzg8L5A=",
        version = "v2.0.0-20180818164646-67afb5ed74ec",
    )
    go_repository(
        name = "com_github_antihax_optional",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/antihax/optional",
        sum = "h1:uZuxRZCz65cG1o6K/xUqImNcYKtmk9ylqaH0itMSvzA=",
        version = "v0.0.0-20180407024304-ca021399b1a6",
    )
    go_repository(
        name = "com_github_yuin_goldmark",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/yuin/goldmark",
        sum = "h1:nqDD4MMMQA0lmWq03Z2/myGPYLQoXtmi0rGVs95ntbo=",
        version = "v1.1.27",
    )
    go_repository(
        name = "com_github_abourget_teamcity",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/abourget/teamcity",
        replace = "github.com/cockroachdb/teamcity",
        sum = "h1:UAqRo5xPCyTtZznAJ9iPVpDUFxGI0a6QWtQ8E+zwJRg=",
        version = "v0.0.0-20180905144921-8ca25c33eb11",
    )
    go_repository(
        name = "com_github_alexbrainman_sspi",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/alexbrainman/sspi",
        sum = "h1:P5U+E4x5OkVEKQDklVPmzs71WM56RTTRqV4OrDC//Y4=",
        version = "v0.0.0-20180613141037-e580b900e9f5",
    )
    go_repository(
        name = "com_github_andy_kimball_arenaskl",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/andy-kimball/arenaskl",
        sum = "h1:vCvyXiLsgAs7qgclk56iBTJQ+gdfiVuzfe5T6sVBL+w=",
        version = "v0.0.0-20200617143215-f701008588b9",
    )
    go_repository(
        name = "com_github_andybalholm_cascadia",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/andybalholm/cascadia",
        sum = "h1:vuRCkM5Ozh/BfmsaTm26kbjm0mIOM3yS5Ek/F5h18aE=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_aokoli_goutils",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/aokoli/goutils",
        sum = "h1:7fpzNGoJ3VA8qcrm++XEE1QUe0mIwNeLa02Nwq7RDkg=",
        version = "v1.0.1",
    )
    go_repository(
        name = "com_github_apache_arrow_go_arrow",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/apache/arrow/go/arrow",
        sum = "h1:kLPoYgtEyqP5M5o1H+oAe5ZjOrL4LLo7jwF9W4hnNq8=",
        version = "v0.0.0-20200610220642-670890229854",
    )
    go_repository(
        name = "com_github_armon_circbuf",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/armon/circbuf",
        sum = "h1:QEF07wC0T1rKkctt1RINW/+RMTVmiwxETico2l3gxJA=",
        version = "v0.0.0-20150827004946-bbbad097214e",
    )
    go_repository(
        name = "com_github_axiomhq_hyperloglog",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/axiomhq/hyperloglog",
        sum = "h1:190ugM9MsyFauTkR/UqcHG/mn5nmFe6SvHJqEHIrtrA=",
        version = "v0.0.0-20181223111420-4b99d0c2c99e",
    )
    go_repository(
        name = "com_github_azure_go_autorest_autorest_azure_auth",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/Azure/go-autorest/autorest/azure/auth",
        sum = "h1:iM6UAvjR97ZIeR93qTcwpKNMpV+/FTWjwEbuPD495Tk=",
        version = "v0.4.2",
    )
    go_repository(
        name = "com_github_azure_go_autorest_autorest_azure_cli",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/Azure/go-autorest/autorest/azure/cli",
        sum = "h1:LXl088ZQlP0SBppGFsRZonW6hSvwgL5gRByMbvUbx8U=",
        version = "v0.3.1",
    )
    go_repository(
        name = "com_github_benesch_cgosymbolizer",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/benesch/cgosymbolizer",
        sum = "h1:Llg88pHOiUbcFOFhr009G8fOBQL6gaVDnxUUeWZuLog=",
        version = "v0.0.0-20180702220239-70e1ee2b39d3",
    )
    go_repository(
        name = "com_github_biogo_store",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/biogo/store",
        sum = "h1:tYoz1OeRpx3dJZlh9T4dQt4kAndcmpl+VNdzbSgFC/0=",
        version = "v0.0.0-20160505134755-913427a1d5e8",
    )
    go_repository(
        name = "com_github_broady_gogeohash",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/broady/gogeohash",
        sum = "h1:iEdmkrNMLXbM7ecffOAtZJQOQUTE4iMonxrb5opUgE4=",
        version = "v0.0.0-20120525094510-7b2c40d64042",
    )
    go_repository(
        name = "com_github_cenkalti_backoff_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/cenkalti/backoff/v3",
        sum = "h1:ske+9nBpD9qZsTBoF41nW5L+AIuFBKMeze18XQ3eG1c=",
        version = "v3.0.0",
    )
    go_repository(
        name = "com_github_cockroachdb_apd_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/cockroachdb/apd/v2",
        sum = "h1:weh8u7Cneje73dDh+2tEVLUvyBc89iwepWCD8b8034E=",
        version = "v2.0.2",
    )
    go_repository(
        name = "com_github_cockroachdb_circuitbreaker",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/cockroachdb/circuitbreaker",
        sum = "h1:u3uQ4oAKM5g2eODBAsDdDSrTs7zRWXtvu+nvSDA9098=",
        version = "v2.2.2-0.20190114160014-a614b14ccf63+incompatible",
    )
    go_repository(
        name = "com_github_cockroachdb_cmux",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/cockroachdb/cmux",
        sum = "h1:dzj1/xcivGjNPwwifh/dWTczkwcuqsXXFHY1X/TZMtw=",
        version = "v0.0.0-20170110192607-30d10be49292",
    )
    go_repository(
        name = "com_github_cockroachdb_cockroach",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/cockroachdb/cockroach",
        sum = "h1:jm8WibQ2uQCc8V/wKkax3XpH6rpmiM1H7J+MZG5yxUM=",
        version = "v0.0.0-20210204185734-162c5ac4968c",
    )
    go_repository(
        name = "com_github_cockroachdb_cockroach_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/cockroachdb/cockroach-go",
        sum = "h1:eqJbq0A8ev6101p/zV72eOM+Z3WZkgb66S7PVJQR9wI=",
        version = "v0.0.0-20200504194139-73ffeee90b62",
    )
    go_repository(
        name = "com_github_cockroachdb_crlfmt",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/cockroachdb/crlfmt",
        sum = "h1:O4yb+RtkmDaXFlHSBpZodXaghfTd3Prdea0nk0eZCT4=",
        version = "v0.0.0-20200116191136-a78e1c207bc0",
    )
    go_repository(
        name = "com_github_cockroachdb_go_test_teamcity",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/cockroachdb/go-test-teamcity",
        sum = "h1:YqzBA7tf8Gv8Oz0BbBsPenqkyjiohS7EUIwi7p1QJCU=",
        version = "v0.0.0-20191211140407-cff980ad0a55",
    )
    go_repository(
        name = "com_github_cockroachdb_gostdlib",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/cockroachdb/gostdlib",
        sum = "h1:TzSEPYgkKDNei3gbLc0rrHu4iHyBp7/+NxPOFmcXGaw=",
        version = "v1.13.0",
    )
    go_repository(
        name = "com_github_cockroachdb_pebble",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/cockroachdb/pebble",
        sum = "h1:ftr0cgVMzdHtmUyeWGZGue2gjVrIdJm7M1DeC9MIecA=",
        version = "v0.0.0-20201228155439-c3ef93f9a9ed",
    )
    go_repository(
        name = "com_github_cockroachdb_returncheck",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/cockroachdb/returncheck",
        sum = "h1:KFOt5I9nEKZgCnOSmy8r4Oykh8BYQO8bFOTgHDS8YZA=",
        version = "v0.0.0-20200612231554-92cdbca611dd",
    )
    go_repository(
        name = "com_github_cockroachdb_stress",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/cockroachdb/stress",
        sum = "h1:7bnDlMxIJCg3o3vIILG2COsbNZpiYHgI4UkCYmeAWnQ=",
        version = "v0.0.0-20170808184505-29b5d31b4c3a",
    )
    go_repository(
        name = "com_github_cockroachdb_ttycolor",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/cockroachdb/ttycolor",
        sum = "h1:TNsiMS1Ij2aleP/05UNSPzsOu9eJm9mfUGLm7Ylt7Dg=",
        version = "v0.0.0-20180709150743-a1d5aaeb377d",
    )
    go_repository(
        name = "com_github_codahale_hdrhistogram",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/codahale/hdrhistogram",
        sum = "h1:qMd81Ts1T2OTKmB4acZcyKaMtRnY5Y44NuXGX2GFJ1w=",
        version = "v0.0.0-20161010025455-3a0bb77429bd",
    )
    go_repository(
        name = "com_github_codefor_geohash",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/Codefor/geohash",
        sum = "h1:iG9B49Q218F/XxXNRM7k/vWf7MKmLIS8AcJV9cGN4nA=",
        version = "v0.0.0-20140723084247-1b41c28e3a9d",
    )
    go_repository(
        name = "com_github_data_dog_go_sqlmock",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/DATA-DOG/go-sqlmock",
        sum = "h1:2L2f5t3kKnCLxnClDD/PrDfExFFa1wjESgxHG/B1ibo=",
        version = "v1.3.2",
    )
    go_repository(
        name = "com_github_dave_dst",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/dave/dst",
        sum = "h1:5wtsjxee7nUDlKEz4i6ewJVn1193vfv2UpEXKqzmaUI=",
        version = "v0.24.0",
    )
    go_repository(
        name = "com_github_dave_gopackages",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/dave/gopackages",
        sum = "h1:l99YKCdrK4Lvb/zTupt0GMPfNbncAGf8Cv/t1sYLOg0=",
        version = "v0.0.0-20170318123100-46e7023ec56e",
    )
    go_repository(
        name = "com_github_dave_jennifer",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/dave/jennifer",
        sum = "h1:S15ZkFMRoJ36mGAQgWL1tnr0NQJh9rZ8qatseX/VbBc=",
        version = "v1.2.0",
    )
    go_repository(
        name = "com_github_dave_kerr",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/dave/kerr",
        sum = "h1:xURkGi4RydhyaYR6PzcyHTueQudxY4LgxN1oYEPJHa0=",
        version = "v0.0.0-20170318121727-bc25dd6abe8e",
    )
    go_repository(
        name = "com_github_dave_rebecca",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/dave/rebecca",
        sum = "h1:jxVfdOxRirbXL28vXMvUvJ1in3djwkVKXCq339qhBL0=",
        version = "v0.9.1",
    )
    go_repository(
        name = "com_github_dgryski_go_metro",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/dgryski/go-metro",
        sum = "h1:8WFBn63wegobsYAX0YjD+8suexZDga5CctH4CCTx2+8=",
        version = "v0.0.0-20180109044635-280f6062b5bc",
    )
    go_repository(
        name = "com_github_edsrzf_mmap_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/edsrzf/mmap-go",
        sum = "h1:CEBF7HpRnUCSJgGUb5h1Gm7e3VkmVDrR8lvWVLtrOFw=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_elastic_gosigar",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/elastic/gosigar",
        sum = "h1:bPIzW1Qkut7n9uwvPAXbnLDVEd45TV5ZwxYZAVX/zEQ=",
        version = "v0.10.0",
    )
    go_repository(
        name = "com_github_elazarl_go_bindata_assetfs",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/elazarl/go-bindata-assetfs",
        sum = "h1:G/bYguwHIzWq9ZoyUQqrjTmJbbYn3j3CKKpKinvZLFk=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_facebookgo_clock",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/facebookgo/clock",
        sum = "h1:yDWHCSQ40h88yih2JAcL6Ls/kVkSE8GFACTGVnMPruw=",
        version = "v0.0.0-20150410010913-600d898af40a",
    )
    go_repository(
        name = "com_github_fanixk_geohash",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/fanixk/geohash",
        sum = "h1:Fyfh/dsHFrC6nkX7H7+nFdTd1wROlX/FxEIWVpKYf1U=",
        version = "v0.0.0-20150324002647-c1f9b5fa157a",
    )
    go_repository(
        name = "com_github_ghemawat_stream",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/ghemawat/stream",
        sum = "h1:r5GgOLGbza2wVHRzK7aAj6lWZjfbAwiu/RDCVOKjRyM=",
        version = "v0.0.0-20171120220530-696b145b53b9",
    )
    go_repository(
        name = "com_github_go_gl_glfw_v3_3_glfw",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/go-gl/glfw/v3.3/glfw",
        sum = "h1:WtGNWLvXpe6ZudgnXrq0barxBImvnnJoMEhXAzcbM0I=",
        version = "v0.0.0-20200222043503-6f7a984d4dc4",
    )
    go_repository(
        name = "com_github_go_ole_go_ole",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/go-ole/go-ole",
        sum = "h1:QNWhweRd9D5Py2rRVboZ2L4SEoW/dyraWJCc8bgS8kE=",
        version = "v1.2.2",
    )
    go_repository(
        name = "com_github_golang_commonmark_html",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/golang-commonmark/html",
        sum = "h1:FeNEDxIy7XouGTJKiJ9Ze5vUbcAIW/FRhQbtKBNmEz8=",
        version = "v0.0.0-20180910111043-7d7c804e1d46",
    )
    go_repository(
        name = "com_github_golang_commonmark_linkify",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/golang-commonmark/linkify",
        sum = "h1:TkuRzcq232K5ytXtQ+BPicsjYWZgt/lS6gJ5HqcUifQ=",
        version = "v0.0.0-20180910111149-f05efb453a0e",
    )
    go_repository(
        name = "com_github_golang_commonmark_markdown",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/golang-commonmark/markdown",
        sum = "h1:YaQaotRjMcVth1VzHUEQlD2oeyQAglA7CXdxp9QLvKM=",
        version = "v0.0.0-20180910011815-a8f139058164",
    )
    go_repository(
        name = "com_github_golang_commonmark_mdurl",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/golang-commonmark/mdurl",
        sum = "h1:XkgfhPs5AotQfcu3EfDEjyAUx91KdtjrxHXYGnZJhoU=",
        version = "v0.0.0-20180910110917-8d018c6567d6",
    )
    go_repository(
        name = "com_github_golang_commonmark_puny",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/golang-commonmark/puny",
        sum = "h1:DUgQdQmDg4sk4SfNR+qOkXcopGz36BL02vp/V7WbPQI=",
        version = "v0.0.0-20180910110745-050be392d8b8",
    )
    go_repository(
        name = "com_github_golang_geo",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/golang/geo",
        sum = "h1:C/hKUcHT483btRbeGkrRjJz+Zbcj8audldIi9tRJDCc=",
        version = "v0.0.0-20200319012246-673a6f80352d",
    )
    go_repository(
        name = "com_github_golang_lint",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/golang/lint",
        sum = "h1:2hRPrmiwPrp3fQX967rNJIhQPtiGXdlQWAxKbKw3VHA=",
        version = "v0.0.0-20180702182130-06c8688daad7",
    )
    go_repository(
        name = "com_github_google_flatbuffers",
        build_file_proto_mode = "disable_global",
        importpath = "github.com/google/flatbuffers",
        sum = "h1:O7CEyB8Cb3/DmtxODGtLHcEvpr81Jm5qLg/hsHnxA2A=",
        version = "v1.11.0",
    )
    go_repository(
        name = "com_github_gorhill_cronexpr",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/gorhill/cronexpr",
        sum = "h1:Spo+4PFAGDqULAsZ7J69MOxq4/fwgZ0zvmDTBqpq7yU=",
        version = "v0.0.0-20140423231348-a557574d6c02",
    )
    go_repository(
        name = "com_github_goware_modvendor",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/goware/modvendor",
        sum = "h1:pQoHt7SOUiWSwN/W5FzibTQLx/1Xa3VMBRcZGtdb1wo=",
        version = "v0.3.0",
    )
    go_repository(
        name = "com_github_grpc_ecosystem_grpc_opentracing",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/grpc-ecosystem/grpc-opentracing",
        sum = "h1:MJG/KsmcqMwFAkh8mTnAwhyKoB+sTAnY4CACC110tbU=",
        version = "v0.0.0-20180507213350-8e809c8a8645",
    )
    go_repository(
        name = "com_github_ianlancetaylor_cgosymbolizer",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/ianlancetaylor/cgosymbolizer",
        sum = "h1:7Mr0IG3PMp/Z/iNN2vCCHdXS4RFzBwi9sr5Km6b2+N4=",
        version = "v0.0.0-20170921033129-f5072df9c550",
    )
    go_repository(
        name = "com_github_ianlancetaylor_demangle",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/ianlancetaylor/demangle",
        sum = "h1:UDMh68UUwekSh5iP2OMhRRZJiiBccgV7axzUG8vi56c=",
        version = "v0.0.0-20181102032728-5e5cf60278f6",
    )
    go_repository(
        name = "com_github_jackc_fake",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/jackc/fake",
        sum = "h1:vr3AYkKovP8uR8AvSGGUK1IDqRa5lAAvEkZG1LKaCRc=",
        version = "v0.0.0-20150926172116-812a484cc733",
    )
    go_repository(
        name = "com_github_jackc_pgx",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/jackc/pgx",
        sum = "h1:2zP5OD7kiyR3xzRYMhOcXVvkDZsImVXfj+yIyTQf3/o=",
        version = "v3.6.2+incompatible",
    )
    go_repository(
        name = "com_github_jaegertracing_jaeger",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/jaegertracing/jaeger",
        sum = "h1:SekUYENjk7gp38SJWM7kKuJjZ1EK7dNqE9UjKaCNJD8=",
        version = "v1.17.0",
    )
    go_repository(
        name = "com_github_jcmturner_aescts_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/jcmturner/aescts/v2",
        sum = "h1:9YKLH6ey7H4eDBXW8khjYslgyqG2xZikXP0EQFKrle8=",
        version = "v2.0.0",
    )
    go_repository(
        name = "com_github_jcmturner_dnsutils_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/jcmturner/dnsutils/v2",
        sum = "h1:lltnkeZGL0wILNvrNiVCR6Ro5PGU/SeBvVO/8c/iPbo=",
        version = "v2.0.0",
    )
    go_repository(
        name = "com_github_jcmturner_goidentity_v6",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/jcmturner/goidentity/v6",
        sum = "h1:VKnZd2oEIMorCTsFBnJWbExfNN7yZr3EhJAxwOkZg6o=",
        version = "v6.0.1",
    )
    go_repository(
        name = "com_github_jcmturner_gokrb5_v8",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/jcmturner/gokrb5/v8",
        sum = "h1:lzPl/30ZLkTveYsYZPKMcgXc8MbnE6RsTd4F9KgiLtk=",
        version = "v8.2.0",
    )
    go_repository(
        name = "com_github_jcmturner_rpc_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/jcmturner/rpc/v2",
        sum = "h1:gMB4IwRXYsWw4Bc6o/az2HJgFUA1ffSh90i26ZJ6Xl0=",
        version = "v2.0.2",
    )
    go_repository(
        name = "com_github_jordanlewis_gcassert",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/jordanlewis/gcassert",
        sum = "h1:d1Ov7s7RbMdHcgE1Eh1CxR+yd2TTKYOCHZf92Tm/fLs=",
        version = "v0.0.0-20200706043056-bf61eb72ee48",
    )
    go_repository(
        name = "com_github_kevinburke_go_bindata",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/kevinburke/go-bindata",
        sum = "h1:hThDhUBH4KjTyhfXfOgacEPfFBNjltnzl/xzfLfrPoQ=",
        version = "v3.13.0+incompatible",
    )
    go_repository(
        name = "com_github_knz_go_libedit",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/knz/go-libedit",
        sum = "h1:0pHpWtx9vcvC0xGZqEQlQdfSQs7WRlAjuPvk3fOZDCo=",
        version = "v1.10.1",
    )
    go_repository(
        name = "com_github_knz_strtime",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/knz/strtime",
        sum = "h1:GQE1iatYDRrIidq4Zf/9ZzKWyrTk2sXOYc1JADbkAjQ=",
        version = "v0.0.0-20200318182718-be999391ffa9",
    )
    go_repository(
        name = "com_github_leanovate_gopter",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/leanovate/gopter",
        sum = "h1:+OUpk+IVvmKU0jivOVFGtOzA6U5AWFs8HE4DRzWLOUE=",
        version = "v0.2.5-0.20190402064358-634a59d12406",
    )
    go_repository(
        name = "com_github_lib_pq_auth_kerberos",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/lib/pq/auth/kerberos",
        sum = "h1:q7hY+WNJTcSqJNGwJzXZYL++nWBaoKlKdgZOyY6jxz4=",
        version = "v0.0.0-20200720160335-984a6aa1ca46",
    )
    go_repository(
        name = "com_github_lightstep_lightstep_tracer_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/lightstep/lightstep-tracer-go",
        sum = "h1:D0GGa7afJ7GcQvu5as6ssLEEKYXvRgKI5d5cevtz8r4=",
        version = "v0.15.6",
    )
    go_repository(
        name = "com_github_linkedin_goavro_v2",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/linkedin/goavro/v2",
        sum = "h1:4OGyN6pvutqmVtObCSzm0BQ+wQQNxSuHCN0t8UAfrjk=",
        version = "v2.8.1",
    )
    go_repository(
        name = "com_github_lufia_iostat",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/lufia/iostat",
        sum = "h1:umdPgllno4d03WCBsA0UIwzGIJYV8FfV4325kOYIhBU=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_lusis_go_slackbot",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/lusis/go-slackbot",
        sum = "h1:AsEBgzv3DhuYHI/GiQh2HxvTP71HCCE9E/tzGUzGdtU=",
        version = "v0.0.0-20180109053408-401027ccfef5",
    )
    go_repository(
        name = "com_github_lusis_slack_test",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/lusis/slack-test",
        sum = "h1:MNApn+Z+fIT4NPZopPfCc1obT6aY3SVM6DOctz1A9ZU=",
        version = "v0.0.0-20190426140909-c40012f20018",
    )
    go_repository(
        name = "com_github_maruel_panicparse",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/maruel/panicparse",
        sum = "h1:mtX2D0ta3lWxCvv276VVIH6mMYzm4jhSfYP70ZJSOQU=",
        version = "v1.1.2-0.20180806203336-f20d4c4d746f",
    )
    go_repository(
        name = "com_github_marusama_semaphore",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/marusama/semaphore",
        sum = "h1:sq+a5mb8zHbmHhrIH06oqIMGsanjpbxNgxEgZVfgpvQ=",
        version = "v0.0.0-20190110074507-6952cef993b2",
    )
    go_repository(
        name = "com_github_masterminds_semver",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/Masterminds/semver",
        sum = "h1:H65muMkzWKEuNDnfl9d70GUjFniHKHRbFPGBuZ3QEww=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_github_masterminds_sprig",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/Masterminds/sprig",
        sum = "h1:z4yfnGrZ7netVz+0EDJ0Wi+5VZCSYp4Z0m2dk6cEM60=",
        version = "v2.22.0+incompatible",
    )
    go_repository(
        name = "com_github_mibk_dupl",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/mibk/dupl",
        sum = "h1:aZc3jqrF9n0tUHwHt/+jsRxA8cRgA0Gdl56M7W7PoqE=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_michaeltjones_walk",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/MichaelTJones/walk",
        sum = "h1:FSsoaa1q4jAaeiAUxf9H0PgFP7eA/UL6c3PdJH+nMN4=",
        version = "v0.0.0-20161122175330-4748e29d5718",
    )
    go_repository(
        name = "com_github_mmatczuk_go_generics",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/mmatczuk/go_generics",
        sum = "h1:QjLgi/L+MjxysinrA8KkNZLf3cAhTluBoSXUvFeN144=",
        version = "v0.0.0-20181212143635-0aaa050f9bab",
    )
    go_repository(
        name = "com_github_mmcloughlin_geohash",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/mmcloughlin/geohash",
        sum = "h1:FihR004p/aE1Sju6gcVq5OLDqGcMnpBY+8moBqIsVOs=",
        version = "v0.9.0",
    )
    go_repository(
        name = "com_github_montanaflynn_stats",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/montanaflynn/stats",
        sum = "h1:l4v5f5zMxqzCwPRRIlL5yCD/REWJ94PrId/ESmRPEDc=",
        version = "v0.4.1-0.20190115100425-713f2944833c",
    )
    go_repository(
        name = "com_github_mwitkow_go_proto_validators",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/mwitkow/go-proto-validators",
        sum = "h1:28i1IjGcx8AofiB4N3q5Yls55VEaitzuEPkFJEVgGkA=",
        version = "v0.0.0-20180403085117-0950a7990007",
    )
    go_repository(
        name = "com_github_niemeyer_pretty",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/niemeyer/pretty",
        sum = "h1:fD57ERR4JtEqsWbfPhv4DMiApHyliiK5xCTNVSPiaAs=",
        version = "v0.0.0-20200227124842-a10e7caefd8e",
    )
    go_repository(
        name = "com_github_nlopes_slack",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/nlopes/slack",
        sum = "h1:OVnHm7lv5gGT5gkcHsZAyw++oHVFihbjWbL3UceUpiA=",
        version = "v0.4.0",
    )
    go_repository(
        name = "com_github_nvveen_gotty",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/Nvveen/Gotty",
        sum = "h1:TngWCqHvy9oXAN6lEVMRuU21PR1EtLVZJmdB18Gu3Rw=",
        version = "v0.0.0-20120604004816-cd527374f1e5",
    )
    go_repository(
        name = "com_github_opennota_wd",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/opennota/wd",
        sum = "h1:cVQhwfBgiKTMAdYPbVeuIiTkdY59qZ3sp5RpyO8CNtg=",
        version = "v0.0.0-20180911144301-b446539ab1e7",
    )
    go_repository(
        name = "com_github_opentracing_contrib_go_observer",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/opentracing-contrib/go-observer",
        sum = "h1:lM6RxxfUMrYL/f8bWEUqdXrANWtrL7Nndbm9iFN0DlU=",
        version = "v0.0.0-20170622124052-a52f23424492",
    )
    go_repository(
        name = "com_github_opentracing_opentracing_go",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/opentracing/opentracing-go",
        sum = "h1:pWlfV3Bxv7k65HYwkikxat0+s3pV4bsqf19k25Ur8rU=",
        version = "v1.1.0",
    )
    go_repository(
        name = "com_github_openzipkin_contrib_zipkin_go_opentracing",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/openzipkin-contrib/zipkin-go-opentracing",
        sum = "h1:82Tnq9OJpn+h5xgGpss5/mOv3KXdjtkdorFSOUusjM8=",
        version = "v0.3.5",
    )
    go_repository(
        name = "com_github_ory_dockertest_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/ory/dockertest/v3",
        sum = "h1:I6KNJ6izxGduLACQii2SP/g7GN0JM9Xfaik6aAVaw6Y=",
        version = "v3.6.0",
    )
    go_repository(
        name = "com_github_peterbourgon_g2s",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/peterbourgon/g2s",
        sum = "h1:sKwxy1H95npauwu8vtF95vG/syrL0p8fSZo/XlDg5gk=",
        version = "v0.0.0-20170223122336-d4e7ad98afea",
    )
    go_repository(
        name = "com_github_petermattis_goid",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/petermattis/goid",
        sum = "h1:q2e307iGHPdTGp0hoxKjt1H5pDo6utceo3dQVK3I5XQ=",
        version = "v0.0.0-20180202154549-b0b1615b78e5",
    )
    go_repository(
        name = "com_github_pierrre_compare",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/pierrre/compare",
        sum = "h1:k4IUsHgh+dbcAOIWCfxVa/7G6STjADH2qmhomv+1quc=",
        version = "v1.0.2",
    )
    go_repository(
        name = "com_github_pierrre_geohash",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/pierrre/geohash",
        sum = "h1:f/zfjdV4rVofTCz1FhP07T+EMQAvcMM2ioGZVt+zqjI=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_pkg_browser",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/pkg/browser",
        sum = "h1:49lOXmGaUpV9Fz3gd7TFZY106KVlPVa5jcYD1gaQf98=",
        version = "v0.0.0-20180916011732-0a3d74bf9ce4",
    )
    go_repository(
        name = "com_github_pseudomuto_protoc_gen_doc",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/pseudomuto/protoc-gen-doc",
        sum = "h1:61vWZuxYa8D7Rn4h+2dgoTNqnluBmJya2MgbqO32z6g=",
        version = "v1.3.2",
    )
    go_repository(
        name = "com_github_pseudomuto_protokit",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/pseudomuto/protokit",
        sum = "h1:hlnBDcy3YEDXH7kc9gV+NLaN0cDzhDvD1s7Y6FZ8RpM=",
        version = "v0.2.0",
    )
    go_repository(
        name = "com_github_puerkitobio_goquery",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/PuerkitoBio/goquery",
        sum = "h1:uGvmFXOA73IKluu/F84Xd1tt/z07GYm8X49XKHP7EJk=",
        version = "v1.5.0",
    )
    go_repository(
        name = "com_github_sasha_s_go_deadlock",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/sasha-s/go-deadlock",
        sum = "h1:lMqc+fUb7RrFS3gQLtoQsJ7/6TV/pAIFvBsqX73DK8Y=",
        version = "v0.2.0",
    )
    go_repository(
        name = "com_github_shirou_gopsutil",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/shirou/gopsutil",
        sum = "h1:P37G9YH8M4vqkKcwBosp+URN5O8Tay67D2MbR361ioY=",
        version = "v2.20.6+incompatible",
    )
    go_repository(
        name = "com_github_stackexchange_wmi",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/StackExchange/wmi",
        sum = "h1:fLjPD/aNc3UIOA6tDi6QXUemppXK3P9BI7mr2hd6gx8=",
        version = "v0.0.0-20180116203802-5d049714c4a6",
    )
    go_repository(
        name = "com_github_the42_cartconvert",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/the42/cartconvert",
        sum = "h1:I4DY8wLxJXCrMYzDM6lKCGc3IQwJX0PlTLsd3nQqI3c=",
        version = "v0.0.0-20131203171324-aae784c392b8",
    )
    go_repository(
        name = "com_github_tomihiltunen_geohash_golang",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/TomiHiltunen/geohash-golang",
        sum = "h1:wumPkzt4zaxO4rHPBrjDK8iZMR41C1qs7njNqlacwQg=",
        version = "v0.0.0-20150112065804-b3e4e625abfb",
    )
    go_repository(
        name = "com_github_twpayne_go_geom",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/twpayne/go-geom",
        sum = "h1:O27mIXZnMYiZi0ZD8ewjs/IT/ZOFVbZHBzPjA9skdmg=",
        version = "v1.3.6",
    )
    go_repository(
        name = "com_github_twpayne_go_kml",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/twpayne/go-kml",
        sum = "h1:RI0JKh/VzdK/d+ZxdJzt8Ar921KMYPfg9qkw7vsbAGw=",
        version = "v1.5.1",
    )
    go_repository(
        name = "com_github_twpayne_go_polyline",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/twpayne/go-polyline",
        sum = "h1:EA8HPg0MNS62R5D2E8B6zyz2TMkmkXVlQaVBVgY3F5A=",
        version = "v1.0.0",
    )
    go_repository(
        name = "com_github_twpayne_go_waypoint",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/twpayne/go-waypoint",
        sum = "h1:0DvmyIQBIEYpXJt7Wm242Wn6ePH84FR5wv2Jh/yRgUw=",
        version = "v0.0.0-20200706203930-b263a7f6e4e8",
    )
    go_repository(
        name = "com_github_vividcortex_ewma",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/VividCortex/ewma",
        sum = "h1:MnEK4VOv6n0RSY4vtRe3h11qjxL3+t0B8yOL8iMXdcM=",
        version = "v1.1.1",
    )
    go_repository(
        name = "com_github_wadey_gocovmerge",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/wadey/gocovmerge",
        sum = "h1:W0LEBv82YCGEtcmPA3uNZBI33/qF//HAAs3MawDjRa0=",
        version = "v0.0.0-20160331181800-b5bfa59ec0ad",
    )
    go_repository(
        name = "com_github_zabawaba99_go_gitignore",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "github.com/zabawaba99/go-gitignore",
        sum = "h1:vpcCVk+pSR/6zcurmlGFD3jC5I/7RMl+GwGAPLxvX18=",
        version = "v0.0.0-20200117185801-39e6bddfb292",
    )
    go_repository(
        name = "in_gopkg_jcmturner_goidentity_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "gopkg.in/jcmturner/goidentity.v3",
        sum = "h1:1duIyWiTaYvVx3YX2CYtpJbUFd7/UuPYCfgXtQ3VTbI=",
        version = "v3.0.0",
    )
    go_repository(
        name = "io_vitess_vitess",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "vitess.io/vitess",
        replace = "github.com/cockroachdb/vitess",
        sum = "h1:Ecr6ZKQU8edcSK19d1HJKnQrQyzxCA4heauBJmygepc=",
        version = "v2.2.0-rc.1.0.20180830030426-1740ce8b3188+incompatible",
    )
    go_repository(
        name = "org_golang_x_arch",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "golang.org/x/arch",
        sum = "h1:Vsc61gop4hfHdzQNolo6Fi/sw7TnJ2yl3ZR4i7bYirs=",
        version = "v0.0.0-20180920145803-b19384d3c130",
    )
    go_repository(
        name = "org_golang_x_perf",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "golang.org/x/perf",
        sum = "h1:xYq6+9AtI+xP3M4r0N1hCkHrInHDBohhquRgx9Kk6gI=",
        version = "v0.0.0-20180704124530-6e6d33e29852",
    )
    go_repository(
        name = "tools_gotest_v3",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "gotest.tools/v3",
        sum = "h1:kG1BFyqVHuQoVQiR1bWGnfz/fmHvvuiSPIV7rvl360E=",
        version = "v3.0.2",
    )
