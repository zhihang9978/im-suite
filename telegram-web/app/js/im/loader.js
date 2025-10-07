/**
 * IM-Suite 适配层加载器
 * 负责按正确顺序加载所有适配层模块
 */

/**
 * 模块加载器
 * 确保所有适配层模块按正确顺序加载
 */
class IMModuleLoader {
    constructor() {
        this.modules = [];
        this.loadedModules = new Set();
        this.isLoading = false;
    }

    /**
     * 注册模块
     * @param {string} name - 模块名称
     * @param {string} path - 模块路径
     * @param {Array} dependencies - 依赖模块
     */
    registerModule(name, path, dependencies = []) {
        this.modules.push({
            name,
            path,
            dependencies,
            loaded: false
        });
    }

    /**
     * 加载所有模块
     */
    async loadAllModules() {
        if (this.isLoading) {
            console.log('模块正在加载中...');
            return;
        }

        this.isLoading = true;
        console.log('开始加载 IM-Suite 适配层模块...');

        try {
            // 按依赖关系排序模块
            const sortedModules = this.sortModulesByDependencies();
            
            // 依次加载模块
            for (const module of sortedModules) {
                await this.loadModule(module);
            }

            console.log('所有模块加载完成');
            this.isLoading = false;
            
        } catch (error) {
            console.error('模块加载失败:', error);
            this.isLoading = false;
            throw error;
        }
    }

    /**
     * 按依赖关系排序模块
     */
    sortModulesByDependencies() {
        const sorted = [];
        const visited = new Set();
        const visiting = new Set();

        const visit = (module) => {
            if (visiting.has(module.name)) {
                throw new Error(`循环依赖检测: ${module.name}`);
            }
            if (visited.has(module.name)) {
                return;
            }

            visiting.add(module.name);

            // 先加载依赖模块
            for (const depName of module.dependencies) {
                const depModule = this.modules.find(m => m.name === depName);
                if (depModule) {
                    visit(depModule);
                }
            }

            visiting.delete(module.name);
            visited.add(module.name);
            sorted.push(module);
        };

        for (const module of this.modules) {
            visit(module);
        }

        return sorted;
    }

    /**
     * 加载单个模块
     */
    async loadModule(module) {
        if (this.loadedModules.has(module.name)) {
            console.log(`模块 ${module.name} 已加载`);
            return;
        }

        console.log(`正在加载模块: ${module.name}`);
        
        try {
            await this.loadScript(module.path);
            this.loadedModules.add(module.name);
            module.loaded = true;
            console.log(`模块 ${module.name} 加载成功`);
        } catch (error) {
            console.error(`模块 ${module.name} 加载失败:`, error);
            throw error;
        }
    }

    /**
     * 加载脚本文件
     */
    loadScript(src) {
        return new Promise((resolve, reject) => {
            // 检查脚本是否已存在
            const existingScript = document.querySelector(`script[src="${src}"]`);
            if (existingScript) {
                resolve();
                return;
            }

            const script = document.createElement('script');
            script.src = src;
            script.async = false; // 确保按顺序加载
            
            script.onload = () => {
                console.log(`脚本加载成功: ${src}`);
                resolve();
            };
            
            script.onerror = () => {
                console.error(`脚本加载失败: ${src}`);
                reject(new Error(`无法加载脚本: ${src}`));
            };
            
            document.head.appendChild(script);
        });
    }

    /**
     * 检查模块是否已加载
     */
    isModuleLoaded(name) {
        return this.loadedModules.has(name);
    }

    /**
     * 获取已加载的模块列表
     */
    getLoadedModules() {
        return Array.from(this.loadedModules);
    }
}

// 创建全局模块加载器
const moduleLoader = new IMModuleLoader();

// 注册所有适配层模块
moduleLoader.registerModule('api', '/app/js/im/adapter/api.js');
moduleLoader.registerModule('ws', '/app/js/im/adapter/ws.js', ['api']);
moduleLoader.registerModule('map', '/app/js/im/adapter/map.js', ['api', 'ws']);
moduleLoader.registerModule('debug', '/app/js/im/debug/TestPage.js', ['api', 'ws']);
moduleLoader.registerModule('integration', '/app/js/im/integration.js', ['api', 'ws', 'map', 'debug']);
moduleLoader.registerModule('test', '/app/js/im/test/integration-test.js', ['integration']);

// 页面加载完成后开始加载模块
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => {
        moduleLoader.loadAllModules().catch(error => {
            console.error('IM-Suite 适配层加载失败:', error);
        });
    });
} else {
    // 页面已加载完成，立即开始加载
    moduleLoader.loadAllModules().catch(error => {
        console.error('IM-Suite 适配层加载失败:', error);
    });
}

// 导出模块加载器
window.IMModuleLoader = moduleLoader;

// 导出供其他模块使用
if (typeof module !== 'undefined' && module.exports) {
    module.exports = IMModuleLoader;
}
