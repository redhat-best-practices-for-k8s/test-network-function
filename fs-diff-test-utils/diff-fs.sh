result=`podman diff $1 | cut -d " " -f 2 | grep -E "(^/var/lib/rpm)|(^/var/lib/dpkg)|(^/bin)|(^/sbin)|(^/lib)|(^/lib64)|(^/usr/bin)|(^/usr/sbin)|(^/usr/lib)|(^/usr/lib64)"`
if [ -z "${result}" ]
then
echo empty
else
echo $result
fi