#! /bin/bash

#images=( "debian" "ubuntu" "centos" "alpine" "nginx" "redis" "mysql" "busybox" "mongo" "httpd" "postgres")
#for i in "${images[@]}"
#do
  #docker pull $i
#done


base=("ubuntu" "debian" "centos" "centos" "alpine" "nginx" "redis" "mysql" "busybox" "mongo" "httpd" "postgres" )
tag=("flavor_build_1" "flavor_build_2" "flavor_build_3" "flavor_release_20.02.2017"
 "flavor_build_4" "faultytag" "flavor_build_5" "taste_build_1" "flavor_build_6" "taste_build_2" "taste_build_3" "taste_build_4")

for (( c=0; c<${#base[@]}; c++ )) {
  docker tag ${base[$c]} 127.0.0.1:5000/testrepo:${tag[$c]}
  docker push 127.0.0.1:5000/testrepo:${tag[$c]}
}


# ubuntu flavor_build_1
# debian flavor_build_2
# centos flavor_build_3
# alpine flavor_build_4
# redis flavor_build_5
# busybox flavor_build_6
# centos flavor_release_20.02.2017
# nginx faultytag
# mysql taste_build_1
# mongo taste_build_2
# apache taste_build_3
# postgres taste_build_4
