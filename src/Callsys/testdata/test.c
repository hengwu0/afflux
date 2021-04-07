#include <stdio.h>
#include <pthread.h>
#include <stdlib.h>
#include <unistd.h>
#include <sys/syscall.h>

void *thread_one()
{
    int pid = syscall(SYS_gettid);
    if (fork()==0)
    {
        sleep(1);
        printf("=============child=================\n");
    }else  
    {
        printf("=============parent=================\n");
        printf("thread_child:int %d main process, tid=%ld\n",getpid(),pid);
    }
    int tid = syscall(SYS_gettid);
    char tmp[100]={0};
    sprintf(tmp, "cat /proc/%d/status",  tid);
    system(tmp);

}

void *thread_two()
{
    int pid = syscall(SYS_gettid);
    printf("thread father:pid=%d tid=%d\n",getpid(),pid);
}

int main(int argc, char *argv[])
{
    pid_t pid;
    pthread_t tid_one,tid_two;
    if((pid=fork())==-1)
    {
        perror("fork");
        exit(EXIT_FAILURE);
    }
    else if(pid==0)
    {
        pthread_create(&tid_one,NULL,(void *)thread_one,NULL);
        pthread_join(tid_one,NULL);
    }
    else
    {
        pthread_create(&tid_two,NULL,(void *)thread_two,NULL);
        pthread_join(tid_two,NULL);
    }
    wait(NULL);
    return 0;
}