#1. Найдите полный хеш и комментарий коммита, хеш которого начинается на aefea:
git show aefea

#commit aefead2207ef7e2aa5dc81a34aedf0cad4c32545
#Author: Alisdair McDiarmid <alisdair@users.noreply.github.com>
#Date:   Thu Jun 18 10:29:58 2020 -0400

#    Update CHANGELOG.md

#2. Какому тегу соответствует коммит 85024d3?
git show 85024d3
commit 85024d3100126de36331c6982bfaac02cdab9e76 (tag: v0.12.23) - искомый тег

#3. Сколько родителей у коммита b8d720? Напишите их хеши.

git show 'b8d720^1'
commit 56cd7859e05c36c06b56d013b55a252d0bb7e158
Merge: 58dcac4b7 ffbcf5581
Author: Chris Griggs <cgriggs@hashicorp.com>
Date:   Mon Jan 13 13:19:09 2020 -0800

    Merge pull request #23857 from hashicorp/cgriggs01-stable

    [cherry-pick]add checkpoint links

git show 'b8d720^2'
commit 9ea88f22fc6269854151c571162c5bcf958bee2b
Author: Chris Griggs <cgriggs@hashicorp.com>
Date:   Tue Jan 21 17:08:06 2020 -0800

    add/update community provider listings
git show 'b8d720^3'
fatal: неоднозначный аргумент «b8d720^3»: неизвестная редакция или не путь в рабочем каталоге.

Родителей: 2 т.к. это merge коммит. 
Хеши: 56cd7859e05c36c06b56d013b55a252d0bb7e158 и 9ea88f22fc6269854151c571162c5bcf958bee2b

