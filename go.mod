module github.com/viliproject/vili

require (
	github.com/BurntSushi/toml v0.0.0-20150501104042-056c9bc7be71
	github.com/CloudCom/firego v0.0.0-20151116200822-162d2c933012
	github.com/GeertJohan/go.rice v0.0.0-20151116220442-53da841dfb99
	github.com/PuerkitoBio/purell v0.0.0-20170917143911-fd18e053af8a
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578
	github.com/Sirupsen/logrus v0.0.0-20151123081515-cdaedc68f289
	github.com/armon/consul-api v0.0.0-20150107205647-dcfedd50ed53
	github.com/asaskevich/govalidator v0.0.0-20151108185501-edd46cdac249
	github.com/aws/aws-sdk-go v0.0.0-20180426212155-a16d7fcf892d
	github.com/beevik/etree v0.0.0-20180426053354-4d4283b0f9e5
	github.com/codegangsta/negroni v0.0.0-20150319171304-c7477ad8e330
	github.com/coreos/go-etcd v0.0.0-20151026160318-003851be7bb0
	github.com/crewjam/saml v0.0.0-20180316200432-e231b7a1204a
	github.com/daaku/go.zipexe v0.0.0-20150329023125-a5fe2436ffcb
	github.com/davecgh/go-spew v0.0.0-20151105211317-5215b55f46b2
	github.com/dchest/uniuri v0.0.0-20160212164326-8902c56451e9
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/docker/distribution v0.0.0-20160629222834-b49f8ed894f1
	github.com/emicklei/go-restful v0.0.0-20171005045149-dc0f94ee75de
	github.com/emicklei/go-restful-openapi v0.0.0-20171012193702-0d037b269a5b
	github.com/emicklei/go-restful-swagger12 v0.0.0-20170926063155-7524189396c6
	github.com/facebookgo/clock v0.0.0-20150410010913-600d898af40a
	github.com/facebookgo/grace v0.0.0-20150807214931-053ab5d25436
	github.com/facebookgo/httpdown v0.0.0-20150904193243-1fa03998d201
	github.com/facebookgo/stats v0.0.0-20151006221625-1b76add642e4
	github.com/fsnotify/fsnotify v0.0.0-20170329110642-4da3e2cfbabc
	github.com/ghodss/yaml v0.0.0-20150909031657-73d445a93680
	github.com/go-ini/ini v1.40.0 // indirect
	github.com/go-openapi/jsonpointer v0.0.0-20170102174223-779f45308c19
	github.com/go-openapi/jsonreference v0.0.0-20161105162150-36d33bfe519e
	github.com/go-openapi/spec v0.0.0-20170928160009-48c2a7185575
	github.com/go-openapi/swag v0.0.0-20170606142751-f3f9494671f9
	github.com/go-sql-driver/mysql v0.0.0-20180413181557-3287d94d4c6a
	github.com/gogo/protobuf v1.2.0 // indirect
	github.com/goji/param v0.0.0-20160927210335-d7f49fd7d1ed
	github.com/golang/glog v0.0.0-20150731225221-fca8c8854093
	github.com/golang/protobuf v0.0.0-20160614223140-0c1f6d65b5a1
	github.com/google/btree v0.0.0-20180813153112-4030bb1f1f0c // indirect
	github.com/google/go-github v0.0.0-20170425170326-e8d46665e050
	github.com/google/go-querystring v0.0.0-20151028211038-2a60fc2ba6c1
	github.com/google/gofuzz v0.0.0-20170612174753-24818f796faf
	github.com/googleapis/gnostic v0.2.0 // indirect
	github.com/gopherjs/gopherjs v0.0.0-20160627014742-6a1c576b7adb
	github.com/gorilla/mux v1.6.2 // indirect
	github.com/gorilla/schema v0.0.0-20170612215433-8b1100835db5
	github.com/gregjones/httpcache v0.0.0-20181110185634-c63ab54fda8f // indirect
	github.com/gucumber/gucumber v0.0.0-20180127021336-7d5c79e832a2
	github.com/hashicorp/hcl v0.0.0-20170509225359-392dba7d905e
	github.com/howeyc/gopass v0.0.0-20170109162249-bf9dde6d0d2c // indirect
	github.com/imdario/mergo v0.3.6 // indirect
	github.com/jmespath/go-jmespath v0.0.0-20180206201540-c2b33e8439af // indirect
	github.com/jonboulle/clockwork v0.0.0-20160907122059-bcac9884e750
	github.com/json-iterator/go v1.1.5 // indirect
	github.com/jszwedko/go-circleci v0.0.0-20180714210311-2ec72e1aa0f8
	github.com/jtolds/gls v4.2.0+incompatible
	github.com/juju/ratelimit v1.0.1 // indirect
	github.com/kardianos/osext v0.0.0-20151124170342-10da29423eb9
	github.com/klauspost/compress v0.0.0-20160421081654-14eb9c495119
	github.com/klauspost/cpuid v0.0.0-20160302075316-09cded8978dc
	github.com/klauspost/crc32 v0.0.0-20160219142609-19b0b332c9e4
	github.com/kr/fs v0.0.0-20131111012553-2788f0dbd169
	github.com/kr/pretty v0.0.0-20150520163514-e6ac2fc51e89
	github.com/kr/pty v0.0.0-20151007230424-f7ee69f31298
	github.com/kr/text v0.0.0-20150905224508-bb797dc4fb83
	github.com/labstack/echo v0.0.0-20180416172039-f867058e3ba4
	github.com/labstack/gommon v0.0.0-20180426014445-588f4e8bddc6
	github.com/lsegal/gucumber v0.0.0-20160627200406-090009ec4b32
	github.com/magiconair/properties v0.0.0-20151021204130-6ac0b95f4492
	github.com/mailru/easyjson v0.0.0-20170902151237-2a92e673c9a6
	github.com/mattn/go-colorable v0.0.0-20151110094335-51a7e7a8b166
	github.com/mattn/go-isatty v0.0.0-20151107153648-d6aaa2f596ae
	github.com/mitchellh/mapstructure v0.0.0-20150717051158-281073eb9eb0
	github.com/mjibson/appstats v0.0.0-20151004071057-0542d5f0e87e
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/nlopes/slack v0.0.0-20151115172036-9153359e4c6e
	github.com/nu7hatch/gouuid v0.0.0-20131221200532-179d4d0c4d8d
	github.com/pelletier/go-buffruneio v0.2.0
	github.com/pelletier/go-toml v0.0.0-20170602065532-fe7536c3dee2
	github.com/petar/GoLLRB v0.0.0-20130427215148-53be0d36a84c
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/pkg/errors v0.0.0-20170505043639-c605e284fe17
	github.com/pkg/sftp v0.0.0-20170511000041-a5f8514e29e9
	github.com/pmezard/go-difflib v0.0.0-20151028094244-d8ed2627bdf0
	github.com/rs/cors v0.0.0-20151030053720-ceb1fbf238d7
	github.com/rs/xhandler v0.0.0-20151117192928-768e938dd1b2
	github.com/russellhaering/goxmldsig v0.0.0-20180122054445-a348271703b2
	github.com/shiena/ansicolor v0.0.0-20151119151921-a422bbe96644
	github.com/smartystreets/assertions v0.0.0-20160422195351-40711f774818
	github.com/smartystreets/goconvey v0.0.0-20160523153147-c53abc99456f
	github.com/spf13/afero v0.0.0-20170217164146-9be650865eab
	github.com/spf13/cast v1.1.0
	github.com/spf13/jwalterweatherman v0.0.0-20151106170057-c2aa07df5938
	github.com/spf13/pflag v1.0.4-0.20181223182923-24fa6976df40
	github.com/spf13/viper v0.0.0-20151110042204-e37b56e207dd
	github.com/stretchr/testify v0.0.0-20160615092844-d77da356e56a
	github.com/thoas/stats v0.0.0-20150511071935-54ed61c2b47e
	github.com/tylerb/graceful v1.2.3
	github.com/ugorji/go v0.0.0-20151120143108-ea9cd21fa0bc
	github.com/valyala/bytebufferpool v0.0.0-20160712061250-8ebd0474e5a2
	github.com/valyala/fasthttp v0.0.0-20160718152503-45697fe30a13
	github.com/valyala/fasttemplate v0.0.0-20160315193134-3b874956e03f
	github.com/volkangurel/viper v0.0.0-20170417183942-2106167bed5d
	github.com/xordataexchange/crypt v0.0.0-20150523230031-749e360c8f23
	github.com/zenazn/goji v0.0.0-20180313033536-4a0a089f56df
	golang.org/x/crypto v0.0.0-20181203042331-505ab145d0a9
	golang.org/x/net v0.0.0-20151121034339-4f2fc6c1e69d
	golang.org/x/oauth2 v0.0.0-20151117210313-442624c9ec92
	golang.org/x/sys v0.0.0-20181228144115-9a3f9b0469bb
	golang.org/x/text v0.0.0-20171013141220-c01e4764d870
	google.golang.org/appengine v0.0.0-20171011215012-a2e0dc829727
	google.golang.org/cloud v0.0.0-20151119220103-975617b05ea8
	gopkg.in/airbrake/gobrake.v2 v2.0.2
	gopkg.in/bsm/ratelimit.v1 v1.0.0-20150417081222-bda20d5067a0
	gopkg.in/check.v1 v1.0.0-20161208181325-20d25e280405
	gopkg.in/fsnotify.v1 v1.2.5
	gopkg.in/gemnasium/logrus-airbrake-hook.v2 v2.0.0
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/redis.v3 v3.2.16
	gopkg.in/vmihailenco/msgpack.v2 v2.9.1
	gopkg.in/yaml.v2 v2.0.0-20150924142314-53feefa2559f
	k8s.io/api v0.0.0-20170921200349-81aa34336d28
	k8s.io/apimachinery v0.0.0-20170921165650-3b05bbfa0a45
	k8s.io/client-go v0.0.0-20170922112243-82aa063804cf
	k8s.io/kube-openapi v0.0.0-20170906091745-abfc5fbe1cf8
)
